package bluetooth

type hciInquiryRequest struct {
	deviceID          uint16
	flags             uint16
	lap               [3]uint8
	length            uint8
	numberOfResponses uint8
	response          [255]inquiryInfo
}

type inquiryInfo struct {
	bluetoothDeviceAddress [6]uint8
	scanRepMode            uint8
	scanPeriodMode         uint8
	scanMode               uint8
	deviceClass            [3]uint8
	clockOffset            uint16
}

type remoteNameRequest struct {
	bluetoothDeviceAddress [6]uint8
	scanRepMode            uint8
	scanMode               uint8
	clockOffset            uint16
}

type remoteNameResponse struct {
	status                 uint8
	bluetoothDeviceAddress [6]uint8
	name                   [248]uint8
}

type hciRequest struct {
	ogf         uint16
	ocf         uint16
	event       int32
	request     *remoteNameRequest
	requestLen  int32
	response    *remoteNameResponse
	responseLen int32
}

type hciFilter struct {
	typeMask  uint32
	eventMask [2]uint32
	opcode    uint16
}

type hciEventHdr struct {
	event uint8
	pLen  uint8
}

type hciCommandHdr struct {
	opcode uint16
	pLen   uint8
}

type socklen uint32
