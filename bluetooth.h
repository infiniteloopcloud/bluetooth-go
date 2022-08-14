#import <Foundation/Foundation.h>
#import <CoreBluetooth/CoreBluetooth.h>
@class CustomCBCentralManagerDelegate;

@interface CustomCBCentralManagerDelegate : NSObject <CBCentralManagerDelegate, CBPeripheralManagerDelegate, CBPeripheralDelegate>

@property (atomic) Boolean connecting;
@property (atomic) Boolean connected;
@property (atomic) Boolean running;
@property (readonly) Boolean verbose;
@property (strong) NSString *peripheralID;
@property (strong) NSArray<NSString *> *characteristicIDs;
@property (strong) CBPeripheral *currentPeripheral;
@property (strong) NSError *error;

- (id) initWithVerbose:(Boolean) verbose;
- (BOOL) connect;
- (void)sendMessage:(char *) msgRaw;
- (char *)readMessage;
- (char *)lastRead;

@end

// Binding functions
CustomCBCentralManagerDelegate* createAdapter(BOOL verbose);
void setPeripheralID(void* delegate, char *id);
void setCharacteristicIDs(void* delegate, char **ids, int lenIDs);
BOOL _connect(void* delegate);
void _disconnect(void* delegate);
const char* getPeripheralID(void* delegate);
bool running(void* delegate);
bool connected(void* delegate);
void sendMessage(void* delegate, char *msg);
char* readMessage(void* delegate);
char* lastRead(void* delegate);
