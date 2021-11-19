package catalog

type Schema struct {
	Interfaces  Interfaces  `json:"interfaces"`
	Controllers Controllers `json:"controllers"`
	Doors       Doors       `json:"doors"`
	Cards       Cards       `json:"cards"`
	Groups      Groups      `json:"groups"`
	Events      Events      `json:"events"`
	Logs        Logs        `json:"logs"`
}

type Interfaces struct {
	OID       OID    `json:"OID"`
	Status    Suffix `json:"status"`
	Created   Suffix `json:"created"`
	Deleted   Suffix `json:"deleted"`
	Type      Suffix `json:"type"`
	Name      Suffix `json:"name"`
	Bind      Suffix `json:"bind"`
	Broadcast Suffix `json:"broadcast"`
	Listen    Suffix `json:"listen"`
}

type Controllers struct {
	OID               OID    `json:"OID"`
	Created           Suffix `json:"created"`
	Name              Suffix `json:"name"`
	DeviceID          Suffix `json:"deviceId"`
	Address           Suffix `json:"address"`
	AddressConfigured Suffix `json:"address-configured"`
	AddressStatus     Suffix `json:"address-status"`
	DateTime          Suffix `json:"datetime"`
	DateTimeSystem    Suffix `json:"systemtime"`
	DateTimeStatus    Suffix `json:"datetime-status"`
	Cards             Suffix `json:"cards"`
	CardsStatus       Suffix `json:"cards-status"`
	Events            Suffix `json:"events"`
	EventsStatus      Suffix `json:"events-status"`
	Door1             Suffix `json:"door1"`
	Door2             Suffix `json:"door2"`
	Door3             Suffix `json:"door3"`
	Door4             Suffix `json:"door4"`
}

type Doors struct {
	OID               OID    `json:"OID"`
	Created           Suffix `json:"created"`
	ControllerOID     Suffix `json:"controller-OID"`
	ControllerCreated Suffix `json:"controller-created"`
	ControllerName    Suffix `json:"controller-name"`
	ControllerID      Suffix `json:"controller-ID"`
	ControllerDoor    Suffix `json:"controller-door"`
	Name              Suffix `json:"name"`
	Delay             Suffix `json:"delay"`
	DelayStatus       Suffix `json:"delay-status"`
	DelayConfigured   Suffix `json:"delay-configured"`
	DelayError        Suffix `json:"delay-error"`
	DelayModified     Suffix `json:"delay-modified"`
	Control           Suffix `json:"control"`
	ControlStatus     Suffix `json:"control-status"`
	ControlConfigured Suffix `json:"control-configured"`
	ControlError      Suffix `json:"control-error"`
	ControlModified   Suffix `json:"control-modified"`
	DisplayIndex      Suffix `json:"display-index"`
}

type Cards struct {
	OID     OID    `json:"OID"`
	Created Suffix `json:"created"`
	Name    Suffix `json:"name"`
	Card    Suffix `json:"card"`
	From    Suffix `json:"from"`
	To      Suffix `json:"to"`
	Groups  Suffix `json:"groups"`
}

type Groups struct {
	OID          OID    `json:"OID"`
	Created      Suffix `json:"created"`
	Name         Suffix `json:"name"`
	Doors        Suffix `json:"doors"`
	DisplayIndex Suffix `json:"display-index"`
}

type Events struct {
	OID   OID    `json:"OID"`
	First Suffix `json:"first"`
	Last  Suffix `json:"last"`

	Timestamp  Suffix `json:"timestamp"`
	DeviceID   Suffix `json:"device-id"`
	Index      Suffix `json:"index"`
	Type       Suffix `json:"type"`
	Door       Suffix `json:"door"`
	Direction  Suffix `json:"direction"`
	Card       Suffix `json:"card"`
	Granted    Suffix `json:"granted"`
	Reason     Suffix `json:"reason"`
	DeviceName Suffix `json:"device-name"`
	DoorName   Suffix `json:"door-name"`
	CardName   Suffix `json:"card-name"`
}

type Logs struct {
	OID   OID    `json:"OID"`
	First Suffix `json:"first"`
	Last  Suffix `json:"last"`

	Timestamp Suffix `json:"timestamp"`
	UID       Suffix `json:"uid"`
	Item      Suffix `json:"item"`
	ItemID    Suffix `json:"item-id"`
	ItemName  Suffix `json:"item-name"`
	Field     Suffix `json:"field"`
	Details   Suffix `json:"details"`
}

func GetSchema() Schema {
	return schema
}

var schema = Schema{
	Interfaces: Interfaces{
		OID:       InterfacesOID,
		Status:    InterfaceStatus,
		Created:   InterfaceCreated,
		Deleted:   InterfaceDeleted,
		Type:      InterfaceType,
		Name:      InterfaceName,
		Bind:      LANBindAddress,
		Broadcast: LANBroadcastAddress,
		Listen:    LANListenAddress,
	},

	Controllers: Controllers{
		OID:               ControllersOID,
		Created:           ControllerCreated,
		Name:              ControllerName,
		DeviceID:          ControllerDeviceID,
		Address:           ControllerAddress,
		AddressConfigured: ControllerAddressConfigured,
		AddressStatus:     ControllerAddressStatus,
		DateTime:          ControllerDateTime,
		DateTimeSystem:    ControllerDateTimeSystem,
		DateTimeStatus:    ControllerDateTimeStatus,
		Cards:             ControllerCards,
		CardsStatus:       ControllerCardsStatus,
		Events:            ControllerEvents,
		EventsStatus:      ControllerEventsStatus,
		Door1:             ControllerDoor1,
		Door2:             ControllerDoor2,
		Door3:             ControllerDoor3,
		Door4:             ControllerDoor4,
	},

	Doors: Doors{
		OID:               DoorsOID,
		Created:           DoorCreated,
		ControllerOID:     DoorControllerOID,
		ControllerCreated: DoorControllerCreated,
		ControllerName:    DoorControllerName,
		ControllerID:      DoorControllerID,
		ControllerDoor:    DoorControllerDoor,
		Name:              DoorName,
		Delay:             DoorDelay,
		DelayStatus:       DoorDelayStatus,
		DelayConfigured:   DoorDelayConfigured,
		DelayError:        DoorDelayError,
		DelayModified:     DoorDelayModified,
		Control:           DoorControl,
		ControlStatus:     DoorControlStatus,
		ControlConfigured: DoorControlConfigured,
		ControlError:      DoorControlError,
		ControlModified:   DoorControlModified,
		DisplayIndex:      DoorIndex,
	},

	Cards: Cards{
		OID:     CardsOID,
		Created: CardCreated,
		Name:    CardName,
		Card:    CardNumber,
		From:    CardFrom,
		To:      CardTo,
		Groups:  CardGroups,
	},

	Groups: Groups{
		OID:          GroupsOID,
		Created:      GroupCreated,
		Name:         GroupName,
		Doors:        GroupDoors,
		DisplayIndex: GroupIndex,
	},

	Events: Events{
		OID:   EventsOID,
		First: EventsFirst,
		Last:  EventsLast,

		Timestamp:  EventTimestamp,
		DeviceID:   EventDeviceID,
		Index:      EventIndex,
		Type:       EventType,
		Door:       EventDoor,
		Direction:  EventDirection,
		Card:       EventCard,
		Granted:    EventGranted,
		Reason:     EventReason,
		DeviceName: EventDeviceName,
		DoorName:   EventDoorName,
		CardName:   EventCardName,
	},

	Logs: Logs{
		OID:   LogsOID,
		First: LogsFirst,
		Last:  LogsLast,

		Timestamp: LogTimestamp,
		UID:       LogUID,
		Item:      LogItem,
		ItemID:    LogItemID,
		ItemName:  LogItemName,
		Field:     LogField,
		Details:   LogDetails,
	},
}

const InterfacesOID OID = "0.1"
const ControllersOID OID = "0.2"
const DoorsOID OID = "0.3"
const CardsOID OID = "0.4"
const GroupsOID OID = "0.5"
const EventsOID OID = "0.6"
const LogsOID OID = "0.7"

const InterfaceStatus Suffix = ".0.0"
const InterfaceCreated Suffix = ".0.1"
const InterfaceDeleted Suffix = ".0.2"
const InterfaceType Suffix = ".1"
const InterfaceName Suffix = ".2"
const LANBindAddress Suffix = ".3"
const LANBroadcastAddress Suffix = ".4"
const LANListenAddress Suffix = ".5"

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
const DoorDelayModified = ".2.4"
const DoorControl = ".3"
const DoorControlStatus = ".3.1"
const DoorControlConfigured = ".3.2"
const DoorControlError = ".3.3"
const DoorControlModified = ".3.4"
const DoorIndex Suffix = ".4"

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

const EventsFirst = ".0.1"
const EventsLast = ".0.2"

const EventTimestamp Suffix = ".1"
const EventDeviceID Suffix = ".2"
const EventIndex Suffix = ".3"
const EventType Suffix = ".4"
const EventDoor Suffix = ".5"
const EventDirection Suffix = ".6"
const EventCard Suffix = ".7"
const EventGranted Suffix = ".8"
const EventReason Suffix = ".9"
const EventDeviceName Suffix = ".10"
const EventDoorName Suffix = ".11"
const EventCardName Suffix = ".12"

const LogsFirst = ".0.1"
const LogsLast = ".0.2"

const LogTimestamp Suffix = ".1"
const LogUID Suffix = ".2"
const LogItem Suffix = ".3"
const LogItemID Suffix = ".4"
const LogItemName Suffix = ".5"
const LogField Suffix = ".6"
const LogDetails Suffix = ".7"
