#import <Foundation/Foundation.h>
#import <CoreBluetooth/CoreBluetooth.h>
#include "bluetooth.h"

@implementation CustomCBCentralManagerDelegate {
    Boolean _running;
    Boolean _connecting;
    Boolean _readDone;
    CBCentralManager *manager;
    CBPeripheralManager *peripheralManager;
    CBPeripheral *_currentPeripheral;
    NSMutableDictionary<CBUUID *, CBCharacteristic *> *_writableCharacteristics;
    NSMutableDictionary<CBUUID *, CBCharacteristic *> *_readableCharacteristics;
    NSArray<NSString *> *_characteristicIDs;

    NSString *_peripheralID;
    NSUInteger serviceCount;
    NSError *_error;
}

- (id) init {
    self = [super init];
    if (self) {
        _running = true;
        manager = [[CBCentralManager alloc] initWithDelegate:self queue:nil];
        peripheralManager = [[CBPeripheralManager alloc] initWithDelegate:self queue:nil];
        _writableCharacteristics = [[NSMutableDictionary alloc] init];
        _readableCharacteristics = [[NSMutableDictionary alloc] init];
        _connecting = false;
        serviceCount = 0;
    }
    return self;
}

- (id)initWithVerbose:(Boolean)verbose {
    self = [self init];
    _verbose = verbose;
    return self;
}

- (void)sendMessage:(char *) msgRaw {
    NSString *msg = [[NSString alloc] initWithUTF8String:msgRaw];
    for (CBCharacteristic *c in _writableCharacteristics.allValues) {
        [self verbose:@"Sending message to: [%@]", [c.UUID UUIDString]];
        NSData *d = [[NSData alloc] initWithBytes:msgRaw length:[msg length]];
        [_currentPeripheral writeValue:d forCharacteristic:c type:CBCharacteristicWriteWithoutResponse];

    }
}

// readMessage: Returns a newline (\n) separated string with the read values
- (char *)readMessage {
    for (CBCharacteristic *c in _readableCharacteristics.allValues) {
        _readDone = false;
        [self verbose:@"Reading message from: [%@]", [c.UUID UUIDString]];
        [_currentPeripheral readValueForCharacteristic:c];
        [self waitUntilReadDone];
    }
    NSMutableString *result = [[NSMutableString alloc] init];
    for (CBCharacteristic *c in _readableCharacteristics.allValues) {
        if (c.value && c.value.length != 0) {
            [result appendString:[[NSString alloc] initWithData:c.value encoding:NSASCIIStringEncoding]];
            [result appendString:@"\n"];
        }
    }
    return [result UTF8String];
}

- (void)peripheral:(CBPeripheral *)peripheral
didDiscoverCharacteristicsForService:(CBService *)service
             error:(NSError *)error {
    --serviceCount;
    [self verbose:@"Characteristic discovered for service[%@]", service.UUID];
    for (CBCharacteristic *c in service.characteristics) {
        if (!_characteristicIDs || ([_characteristicIDs containsObject:[c.UUID UUIDString]])) {
            if (!_characteristicIDs) {
                [self info:@"Characteristic id match: %@", [c.UUID UUIDString]];
            }
            [_currentPeripheral setNotifyValue:true forCharacteristic:c];
            if ((c.properties & CBCharacteristicPropertyWriteWithoutResponse) != 0) {
                [self verbose:@"Characteristic[%@] storing writable", c.UUID];
                [_writableCharacteristics setValue:c forKey:[c.UUID UUIDString]];
            }
            if ((c.properties & CBCharacteristicPropertyRead) != 0) {
                [self verbose:@"Characteristic[%@] storing readable", c.UUID];
                [_readableCharacteristics setValue:c forKey:[c.UUID UUIDString]];
            }
        }
    }
    if (error) {
        [self info:@"Unable to discover characteristic service: %@", error];
        [self setError:error];
        return;
    }
    if (serviceCount <= 0) {
        [self setConnected:true];
    }
}

- (void)peripheral:(CBPeripheral *)peripheral didUpdateValueForCharacteristic:(CBCharacteristic *)characteristic error:(NSError *)error {
    [self verbose:@"UpdateValueForCharacteristic done: %@, %@, %@; ERR: %@", peripheral.identifier, [characteristic.UUID UUIDString], characteristic.value, error];
    _readDone = true;
}

- (void)peripheral:(CBPeripheral *)peripheral didWriteValueForCharacteristic:(CBCharacteristic *)characteristic error:(NSError *)error {
    [self verbose:@"WriteValueForCharacteristic done: %@, %@; ERR: %@", peripheral.identifier, [characteristic.UUID UUIDString], error];
}

- (void)centralManager:(CBCentralManager *)central
 didDiscoverPeripheral:(CBPeripheral *)peripheral
     advertisementData:(NSDictionary<NSString *,id> *)advertisementData
                  RSSI:(NSNumber *)RSSI {
    if ((![self connecting]) && ([peripheral.identifier isEqual:[[NSUUID new] initWithUUIDString:self.peripheralID]])) {
        [self info:@"Discovered peripheral[%@] with Name: %@", peripheral.identifier, peripheral.name];
        [self setConnecting:true];
        [self setCurrentPeripheral:peripheral];
        [_currentPeripheral setDelegate:self];
        [manager connectPeripheral:_currentPeripheral options:nil];
    }
}

- (void)centralManager:(CBCentralManager *)central didFailToConnectPeripheral:(CBPeripheral *)peripheral error:(NSError *)error {
    [self info:@"Peripheral[%@] connection failed: %@", peripheral.identifier, error];
    [self setError:error];
}

- (void) centralManager:(CBCentralManager *)central didConnectPeripheral:(CBPeripheral *)peripheral {
    [self info:@"Peripheral[%@] connected", _currentPeripheral.identifier];
    [peripheral discoverServices:nil];
}

- (void)centralManager:(CBCentralManager *)central didDisconnectPeripheral:(CBPeripheral *)peripheral error:(NSError *)error {
    [self info:@"Peripheral[%@] disconnected %@", peripheral.identifier, error];
}

- (void)peripheral:(CBPeripheral *)peripheral didDiscoverServices:(NSError *)error {
    if (error) {
        [self info:@"Discovered service failed: %@", error];
    }
    [self verbose:@"Discovered service count: %lu", peripheral.services.count];
    serviceCount = peripheral.services.count;
    for (CBService *s in peripheral.services) {
        [self verbose:@"Discover service[%@] characteristics", s.UUID];
        [peripheral discoverCharacteristics:nil forService:s];
    }
}

- (void)peripheral:(CBPeripheral *)peripheral didModifyServices:(NSArray<CBService *> *)invalidatedServices {
    for (CBService * s in invalidatedServices) {
        [self verbose:@"Peripheral[%@] service[%@] modified", peripheral.identifier, [s.UUID UUIDString]];
    }
    // TODO add service discovery here
}

- (void)centralManagerDidUpdateState:(nonnull CBCentralManager *)central {
    if (central.state == CBManagerStatePoweredOn) {
        [self verbose:@"Central manager state powered on, start scanning"];
        [central scanForPeripheralsWithServices:nil options:nil];
    } else {
        [self info:@"Central manager state no longer powered on [%lu]", central.state];
    }
}

- (void)peripheralManagerDidUpdateState:(nonnull CBPeripheralManager *)peripheral {
    CBManagerState state = [peripheral state];

    NSString *string = @"Unknown state";

    switch(state)
    {
        case CBManagerStatePoweredOff:
            string = @"CoreBluetooth BLE hardware is powered off";
            break;

        case CBManagerStatePoweredOn:
            string = @"CoreBluetooth BLE hardware is powered on and ready";
            break;

        case CBManagerStateUnauthorized:
            string = @"CoreBluetooth BLE state is unauthorized";
            break;

        case CBManagerStateUnknown:
            string = @"CoreBluetooth BLE state is unknown";
            break;

        case CBManagerStateUnsupported:
            string = @"CoreBluetooth BLE hardware is unsupported on this platform";
            break;

        default:
            break;
    }
    [self info:@"%@", string];
}

- (void)connect {
    NSRunLoop *runLoop = NSRunLoop.currentRunLoop;
    NSDate *distantFuture = NSDate.distantFuture;
    while([self running] && [runLoop runMode:NSDefaultRunLoopMode beforeDate:distantFuture]){
        if (self.connected) {
            [self info:@"Bluetooth successfully connected"];
            break;
        }
    }
}

- (void)waitUntilReadDone {
    NSRunLoop *runLoop = NSRunLoop.currentRunLoop;
    NSDate *distantFuture = NSDate.distantFuture;
    while([runLoop runMode:NSDefaultRunLoopMode beforeDate:distantFuture]){
        if (_readDone) {
            [self verbose:@"Reading process done"];
            break;
        }
    }
}

- (void) verbose:(NSString *) msg, ... {
    if (!_verbose) {
        return;
    }
    va_list args;
    va_start(args, msg);
    [self log:msg withParameters:args];
    va_end(args);
}

- (void) info:(NSString *) msg, ... {
    va_list args;
    va_start(args, msg);
    [self log:msg withParameters:args];
    va_end(args);
}

- (void) log:(NSString *) msg withParameters:(va_list)valist  {
    NSLogv(msg, valist);
}


@end

CustomCBCentralManagerDelegate* createAdapter(BOOL verbose) {
    return [[CustomCBCentralManagerDelegate alloc] initWithVerbose:verbose];
}

void setPeripheralID(void* delegate, char *id) {
    [(CustomCBCentralManagerDelegate*)CFBridgingRelease(delegate) setPeripheralID:[[NSString alloc] initWithUTF8String:id]];
}

void setCharacteristicIDs(void* delegate, char **ids, int lenIDs) {
    NSMutableArray<NSString *> *idArray = [[NSMutableArray<NSString *> alloc] init];
    for (NSInteger i = 0; i < lenIDs; i++) {
        [idArray addObject:[[NSString alloc] initWithUTF8String:ids[i]]];
    }

    [(CustomCBCentralManagerDelegate*)CFBridgingRelease(delegate) setCharacteristicIDs:idArray];
}

void _connect(void* delegate) {
    [(CustomCBCentralManagerDelegate*)CFBridgingRelease(delegate) connect];
}

const char* getPeripheralID(void* delegate) {
    return [[(CustomCBCentralManagerDelegate*)CFBridgingRelease(delegate) peripheralID] UTF8String];
}

bool running(void* delegate) {
    return [(CustomCBCentralManagerDelegate*)CFBridgingRelease(delegate) running];
}

bool connected(void* delegate) {
    return [(CustomCBCentralManagerDelegate*)CFBridgingRelease(delegate) connected];
}

void sendMessage(void* delegate, char *msg) {
    [(CustomCBCentralManagerDelegate*)CFBridgingRelease(delegate) sendMessage:msg];
}

char* readMessage(void* delegate) {
    return [(CustomCBCentralManagerDelegate*)CFBridgingRelease(delegate) readMessage];
}
