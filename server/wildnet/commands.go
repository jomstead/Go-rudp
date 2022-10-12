package wildnet

type Command uint8

const (
	SCAN               Command = iota // SCAN - request a large radial scan of nearby space
	SCAN_RESULT                       // SCAN_RESULT - list of objects in nearby space
	SCAN_TARGET                       // SCAN_TARGET - request details on a single object
	SCAN_TARGET_RESULT                // SCAN_TARGET_RESULT - details for a single object
	MOVE                              // MOVE - Request current ship move to location
	MOVE_RESULT                       // MOVE_RESULT - returns time of arrival for ship
	POSITION                          // POSITION - updated object position
)
