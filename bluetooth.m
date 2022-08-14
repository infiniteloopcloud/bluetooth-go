#import <Foundation/Foundation.h>
#import <CoreBluetooth/CoreBluetooth.h>
#include "bluetooth.h"

@implementation CustomCBCentralManagerDelegate {
    Boolean _running;
    Boolean _connecting;
    Boolean _readDone;
    Boolean _writeDone;
    NSString *_lastRead;
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
        if ((c.properties & CBCharacteristicPropertyWriteWithoutResponse) != 0) {
            [self verbose:@"CBCharacteristicPropertyWriteWithoutResponse"];
            [_currentPeripheral writeValue:d forCharacteristic:c type:CBCharacteristicWriteWithoutResponse];
        } else
            if ((c.properties & CBCharacteristicPropertyWrite) != 0) {
            [self verbose:@"CBCharacteristicPropertyWrite"];
            _writeDone = false;
            _readDone = false;
            [_currentPeripheral writeValue:d forCharacteristic:c type:CBCharacteristicWriteWithResponse];
            [self waitUntilWriteDone];
            [self waitUntilReadDone];
        }
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
        [self verbose:@"CBCharacteristicPropertyRead"];
        if (c.value && c.value.length != 0) {
            [result appendString:[[NSString alloc] initWithData:c.value encoding:NSUTF8StringEncoding]];
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
            [self printProperties:c.properties ID:[c.UUID UUIDString]];
            if ((c.properties & CBCharacteristicPropertyNotify) != 0) {
                [self verbose:@"Characteristic[%@] setting notify", c.UUID];
                [_currentPeripheral setNotifyValue:true forCharacteristic:c];
            }
            if (((c.properties & CBCharacteristicPropertyWriteWithoutResponse) != 0) || ((c.properties & CBCharacteristicPropertyWrite) != 0)) {
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

/*
 CBCharacteristicPropertyBroadcast                                                = 0x01,
 CBCharacteristicPropertyRead                                                    = 0x02,
 CBCharacteristicPropertyWriteWithoutResponse                                    = 0x04,
 CBCharacteristicPropertyWrite                                                    = 0x08,
 CBCharacteristicPropertyNotify                                                    = 0x10,
 CBCharacteristicPropertyIndicate                                                = 0x20,
 CBCharacteristicPropertyAuthenticatedSignedWrites                                = 0x40,
 CBCharacteristicPropertyExtendedProperties                                        = 0x80,
 CBCharacteristicPropertyNotifyEncryptionRequired NS_ENUM_AVAILABLE(10_9, 6_0)    = 0x100,
 CBCharacteristicPropertyIndicateEncryptionRequired NS_ENUM_AVAILABLE(10_9, 6_0)    = 0x200
 */

- (void) printProperties:(CBCharacteristicProperties) props ID:(NSString *) uuid {
    if ((props & CBCharacteristicPropertyBroadcast) != 0) {
        NSLog(@"%@ CBCharacteristicPropertyBroadcast", uuid);
    }
    if ((props & CBCharacteristicPropertyRead) != 0) {
        NSLog(@"%@ CBCharacteristicPropertyRead", uuid);
    }
    if ((props & CBCharacteristicPropertyWriteWithoutResponse) != 0) {
        NSLog(@"%@ CBCharacteristicPropertyWriteWithoutResponse", uuid);
    }
    if ((props & CBCharacteristicPropertyWrite) != 0) {
        NSLog(@"%@ CBCharacteristicPropertyWrite", uuid);
    }
    if ((props & CBCharacteristicPropertyNotify) != 0) {
        NSLog(@"%@ CBCharacteristicPropertyNotify", uuid);
    }
    if ((props & CBCharacteristicPropertyIndicate) != 0) {
        NSLog(@"%@ CBCharacteristicPropertyIndicate", uuid);
    }
    if ((props & CBCharacteristicPropertyAuthenticatedSignedWrites) != 0) {
        NSLog(@"%@ CBCharacteristicPropertyAuthenticatedSignedWrites", uuid);
    }
    if ((props & CBCharacteristicPropertyExtendedProperties) != 0) {
        NSLog(@"%@ CBCharacteristicPropertyExtendedProperties", uuid);
    }
    if ((props & CBCharacteristicPropertyNotifyEncryptionRequired) != 0) {
        NSLog(@"%@ CBCharacteristicPropertyNotifyEncryptionRequired", uuid);
    }
    if ((props & CBCharacteristicPropertyIndicateEncryptionRequired) != 0) {
        NSLog(@"%@ CBCharacteristicPropertyIndicateEncryptionRequired", uuid);
    }
}

- (void)peripheral:(CBPeripheral *)peripheral didUpdateValueForCharacteristic:(CBCharacteristic *)characteristic error:(NSError *)error {
    [self verbose:@"UpdateValueForCharacteristic done: %@, %@, %@; ERR: %@", peripheral.identifier, [characteristic.UUID UUIDString], [[NSString alloc] initWithData:characteristic.value encoding:NSUTF8StringEncoding], error];
    if (characteristic.value) {
        _lastRead = [NSString stringWithUTF8String:[characteristic.value bytes]];
    }
    _readDone = true;
}

- (void)peripheral:(CBPeripheral *)peripheral didWriteValueForCharacteristic:(CBCharacteristic *)characteristic error:(NSError *)error {
    [self verbose:@"WriteValueForCharacteristic done: %@, %@; ERR: %@", peripheral.identifier, [characteristic.UUID UUIDString], error];
    if (error) {
        [self setError:error];
        _readDone = true;
    }
    _writeDone = true;
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
    [self connect];
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

- (BOOL)connect {
    NSRunLoop *runLoop = NSRunLoop.currentRunLoop;
    NSDate *distantFuture = NSDate.distantFuture;
    while([self running] && [runLoop runMode:NSDefaultRunLoopMode beforeDate:distantFuture]){
        if (self.connected) {
            [self info:@"Bluetooth successfully connected"];
            return TRUE;
        }
        if (self.error) {
            [self info:@"Bluetooth connection failed"];
            return FALSE;
        }
    }
    return FALSE;
}

- (void)disconnect {
    [manager cancelPeripheralConnection:self.currentPeripheral];
}

- (char *)lastRead {
    if (_lastRead) {
        return [_lastRead UTF8String];
    }
    return nil;
}

- (void)waitUntilReadDone {
    NSRunLoop *runLoop = NSRunLoop.currentRunLoop;
    NSDate *distantFuture = NSDate.distantFuture;
    while([runLoop runMode:NSDefaultRunLoopMode beforeDate:distantFuture]){
        if (_readDone) {
            [self verbose:@"Read done"];
            break;
        }
    }
}

- (void)waitUntilWriteDone {
    NSRunLoop *runLoop = NSRunLoop.currentRunLoop;
    NSDate *distantFuture = NSDate.distantFuture;
    while([runLoop runMode:NSDefaultRunLoopMode beforeDate:distantFuture]){
        if (_writeDone) {
            [self verbose:@"Write done"];
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
    NSMutableString *mutMsg = [[NSMutableString alloc] initWithUTF8String:"INFO::"];
    [mutMsg appendString:msg];
    [self log:mutMsg withParameters:args];
    va_end(args);
}

- (void) info:(NSString *) msg, ... {
    va_list args;
    va_start(args, msg);
    NSMutableString *mutMsg = [[NSMutableString alloc] initWithUTF8String:"INFO::"];
    [mutMsg appendString:msg];
    [self log:mutMsg withParameters:args];
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

BOOL _connect(void* delegate) {
    return [(CustomCBCentralManagerDelegate*)CFBridgingRelease(delegate) connect];
}

void _disconnect(void* delegate) {
    [(CustomCBCentralManagerDelegate*)CFBridgingRelease(delegate) disconnect];
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

char* lastRead(void* delegate) {
    return [(CustomCBCentralManagerDelegate*)CFBridgingRelease(delegate) lastRead];
}
