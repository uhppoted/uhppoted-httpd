package catalog

// const System OID = "0.1"
// const Doors OID = "0.2"
// const Cards OID = "0.3"
// const Groups OID = "0.4"

const InterfaceType Suffix = ".0"
const InterfaceName Suffix = ".1"
const LANBindAddress Suffix = ".2"
const LANBroadcastAddress Suffix = ".3"
const LANListenAddress Suffix = ".4"

const ControllerCreated = ".0.1"
const ControllerName = ".1"
const ControllerDeviceID = ".2"
const ControllerAddress = ".3"
const ControllerAddressConfigured = ".3.1"
const ControllerAddressStatus = ".3.2"
const ControllerDateTime = ".4"
const ControllerDateTimeSystem = ".4.1"
const ControllerDateTimeStatus = ".4.2"
const ControllerCards = ".5"
const ControllerCardsStatus = ".5.1"
const ControllerEvents = ".6"
const ControllerEventsStatus = ".6.1"
const ControllerDoor1 = ".7"
const ControllerDoor2 = ".8"
const ControllerDoor3 = ".9"
const ControllerDoor4 = ".10"

const CardCreated Suffix = ".0.1"
const CardName Suffix = ".1"
const CardNumber Suffix = ".2"
const CardFrom Suffix = ".3"
const CardTo Suffix = ".4"
const CardGroups Suffix = ".5"

const GroupCreated Suffix = ".0.1"
const GroupName Suffix = ".1"
const GroupDoors Suffix = ".2"
const GroupIndex Suffix = ".3"

const DoorCreated = ".0.1"
const DoorControllerOID = ".0.2"
const DoorControllerCreated = ".0.2.1"
const DoorControllerName = ".0.2.2"
const DoorControllerID = ".0.2.3"
const DoorControllerDoor = ".0.2.4"
const DoorName = ".1"
const DoorDelay = ".2"
const DoorDelayStatus = ".2.1"
const DoorDelayConfigured = ".2.2"
const DoorDelayError = ".2.3"
const DoorControl = ".3"
const DoorControlStatus = ".3.1"
const DoorControlConfigured = ".3.2"
const DoorControlError = ".3.3"
const DoorIndex Suffix = ".4"
