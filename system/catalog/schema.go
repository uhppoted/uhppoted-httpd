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

type Metadata struct {
	Status   Suffix `json:"status"`
	Created  Suffix `json:"created"`
	Deleted  Suffix `json:"deleted"`
	Modified Suffix `json:"modified"`
	Type     Suffix `json:"type"`
}

type Interfaces struct {
	OID OID `json:"OID"`
	Metadata
	Name      Suffix `json:"name"`
	ID        Suffix `json:"ID"`
	Bind      Suffix `json:"bind"`
	Broadcast Suffix `json:"broadcast"`
	Listen    Suffix `json:"listen"`
}

type Controllers struct {
	OID OID `json:"OID"`
	Metadata
	Name              Suffix `json:"name"`
	DeviceID          Suffix `json:"deviceID"`
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
	OID OID `json:"OID"`
	Metadata
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
}

type Cards struct {
	OID OID `json:"OID"`
	Metadata
	Name   Suffix `json:"name"`
	Card   Suffix `json:"card"`
	From   Suffix `json:"from"`
	To     Suffix `json:"to"`
	Groups Suffix `json:"groups"`
}

type Groups struct {
	OID OID `json:"OID"`
	Metadata
	Name  Suffix `json:"name"`
	Doors Suffix `json:"doors"`
}

type Events struct {
	OID OID `json:"OID"`
	Metadata
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
	OID OID `json:"OID"`
	Metadata
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
		OID: InterfacesOID,
		Metadata: Metadata{
			Status:   Status,
			Created:  Created,
			Deleted:  Deleted,
			Modified: Modified,
			Type:     Type,
		},
		Name:      InterfaceName,
		ID:        InterfaceID,
		Bind:      LANBindAddress,
		Broadcast: LANBroadcastAddress,
		Listen:    LANListenAddress,
	},

	Controllers: Controllers{
		OID: ControllersOID,
		Metadata: Metadata{
			Status:   Status,
			Created:  Created,
			Deleted:  Deleted,
			Modified: Modified,
			Type:     Type,
		},
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
		OID: DoorsOID,
		Metadata: Metadata{
			Status:   Status,
			Created:  Created,
			Deleted:  Deleted,
			Modified: Modified,
			Type:     Type,
		},
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
	},

	Cards: Cards{
		OID: CardsOID,
		Metadata: Metadata{
			Status:   Status,
			Created:  Created,
			Deleted:  Deleted,
			Modified: Modified,
			Type:     Type,
		},
		Name:   CardName,
		Card:   CardNumber,
		From:   CardFrom,
		To:     CardTo,
		Groups: CardGroups,
	},

	Groups: Groups{
		OID: GroupsOID,
		Metadata: Metadata{
			Status:   Status,
			Created:  Created,
			Deleted:  Deleted,
			Modified: Modified,
			Type:     Type,
		},
		Name:  GroupName,
		Doors: GroupDoors,
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

const Status Suffix = ".0.0"
const Created Suffix = ".0.1"
const Deleted Suffix = ".0.2"
const Modified Suffix = ".0.3"
const Type Suffix = ".0.4"

const InterfaceName Suffix = ".1"
const InterfaceID Suffix = ".2"
const LANBindAddress Suffix = ".3.1"
const LANBroadcastAddress Suffix = ".3.2"
const LANListenAddress Suffix = ".3.3"

const ControllerName Suffix = ".1"
const ControllerDeviceID Suffix = ".2"
const ControllerAddress Suffix = ".3"
const ControllerAddressConfigured Suffix = ".3.1"
const ControllerAddressStatus Suffix = ".3.2"
const ControllerDateTime Suffix = ".4"
const ControllerDateTimeSystem Suffix = ".4.1"
const ControllerDateTimeStatus Suffix = ".4.2"
const ControllerCards Suffix = ".5"
const ControllerCardsStatus Suffix = ".5.1"
const ControllerEvents Suffix = ".6"
const ControllerEventsStatus Suffix = ".6.1"
const ControllerDoor1 Suffix = ".7"
const ControllerDoor2 Suffix = ".8"
const ControllerDoor3 Suffix = ".9"
const ControllerDoor4 Suffix = ".10"

const DoorControllerOID Suffix = ".0.4"
const DoorControllerCreated Suffix = ".0.4.1"
const DoorControllerName Suffix = ".0.4.2"
const DoorControllerID Suffix = ".0.4.3"
const DoorControllerDoor Suffix = ".0.4.4"
const DoorName Suffix = ".1"
const DoorDelay Suffix = ".2"
const DoorDelayStatus Suffix = ".2.1"
const DoorDelayConfigured Suffix = ".2.2"
const DoorDelayError Suffix = ".2.3"
const DoorDelayModified Suffix = ".2.4"
const DoorControl Suffix = ".3"
const DoorControlStatus Suffix = ".3.1"
const DoorControlConfigured Suffix = ".3.2"
const DoorControlError Suffix = ".3.3"
const DoorControlModified Suffix = ".3.4"

const CardName Suffix = ".1"
const CardNumber Suffix = ".2"
const CardFrom Suffix = ".3"
const CardTo Suffix = ".4"
const CardGroups Suffix = ".5"

const GroupName Suffix = ".1"
const GroupDoors Suffix = ".2"

const EventsFirst Suffix = ".0.1"
const EventsLast Suffix = ".0.2"

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

const LogsFirst Suffix = ".0.1"
const LogsLast Suffix = ".0.2"

const LogTimestamp Suffix = ".1"
const LogUID Suffix = ".2"
const LogItem Suffix = ".3"
const LogItemID Suffix = ".4"
const LogItemName Suffix = ".5"
const LogField Suffix = ".6"
const LogDetails Suffix = ".7"
