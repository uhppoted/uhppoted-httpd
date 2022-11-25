package schema

type Schema struct {
	Interfaces  Interfaces  `json:"interfaces"`
	Controllers Controllers `json:"controllers"`
	Doors       Doors       `json:"doors"`
	Cards       Cards       `json:"cards"`
	Groups      Groups      `json:"groups"`
	Events      Events      `json:"events"`
	Logs        Logs        `json:"logs"`
	Users       Users       `json:"users"`
}

type Metadata struct {
	Status   Suffix `json:"status"`
	Created  Suffix `json:"created"`
	Deleted  Suffix `json:"deleted"`
	Modified Suffix `json:"modified"`
	Type     Suffix `json:"type"`
	Touched  Suffix `json:"touched"`
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
	Name     Suffix `json:"name"`
	DeviceID Suffix `json:"deviceID"`
	Endpoint struct {
		Status     Suffix `json:"status"`
		Address    Suffix `json:"address"`
		Configured Suffix `json:"configured"`
	} `json:"endpoint"`
	DateTime struct {
		Status     Suffix `json:"status"`
		Current    Suffix `json:"datetime"`
		Configured Suffix `json:"configured"`
		Modified   Suffix `json:"modified"`
	} `json:"datetime"`
	Cards struct {
		Status Suffix `json:"status"`
		Count  Suffix `json:"count"`
	} `json:"cards"`
	Events struct {
		Status  Suffix `json:"status"`
		First   Suffix `json:"first"`
		Last    Suffix `json:"last"`
		Current Suffix `json:"current"`
	} `json:"events"`
	Doors struct {
		Door1 Suffix `json:"1"`
		Door2 Suffix `json:"2"`
		Door3 Suffix `json:"3"`
		Door4 Suffix `json:"4"`
	} `json:"doors"`
}

type Doors struct {
	OID OID `json:"OID"`
	Metadata
	Name  Suffix `json:"name"`
	Delay struct {
		Delay      Suffix `json:"delay"`
		Status     Suffix `json:"status"`
		Configured Suffix `json:"configured"`
		Error      Suffix `json:"error"`
		Modified   Suffix `json:"modified"`
	} `json:"delay"`
	Control struct {
		Control    Suffix `json:"control"`
		Status     Suffix `json:"status"`
		Configured Suffix `json:"configured"`
		Error      Suffix `json:"error"`
		Modified   Suffix `json:"modified"`
	} `json:"control"`
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
	Status Suffix `json:"status"`
	First  Suffix `json:"first"`
	Last   Suffix `json:"last"`

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

type Users struct {
	OID OID `json:"OID"`
	Metadata
	Name     Suffix `json:"name"`
	UID      Suffix `json:"uid"`
	Role     Suffix `json:"role"`
	Password Suffix `json:"password"`
	OTP      Suffix `json:"otp"`
	OTPKey   Suffix `json:"otpkey"`
	Locked   Suffix `json:"locked"`
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
		Name:     ControllerName,
		DeviceID: ControllerDeviceID,
		Endpoint: struct {
			Status     Suffix `json:"status"`
			Address    Suffix `json:"address"`
			Configured Suffix `json:"configured"`
		}{
			Status:     ControllerEndpointStatus,
			Address:    ControllerEndpointAddress,
			Configured: ControllerEndpointConfigured,
		},
		DateTime: struct {
			Status     Suffix `json:"status"`
			Current    Suffix `json:"datetime"`
			Configured Suffix `json:"configured"`
			Modified   Suffix `json:"modified"`
		}{
			Status:     ControllerDateTimeStatus,
			Current:    ControllerDateTimeCurrent,
			Configured: ControllerDateTimeConfigured,
			Modified:   ControllerDateTimeModified,
		},
		Cards: struct {
			Status Suffix `json:"status"`
			Count  Suffix `json:"count"`
		}{
			Status: ControllerCardsStatus,
			Count:  ControllerCardsCount,
		},
		Events: struct {
			Status  Suffix `json:"status"`
			First   Suffix `json:"first"`
			Last    Suffix `json:"last"`
			Current Suffix `json:"current"`
		}{
			Status:  ControllerEventsStatus,
			First:   ControllerEventsFirst,
			Last:    ControllerEventsLast,
			Current: ControllerEventsCurrent,
		},
		Doors: struct {
			Door1 Suffix `json:"1"`
			Door2 Suffix `json:"2"`
			Door3 Suffix `json:"3"`
			Door4 Suffix `json:"4"`
		}{
			Door1: ControllerDoor1,
			Door2: ControllerDoor2,
			Door3: ControllerDoor3,
			Door4: ControllerDoor4,
		},
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
		Name: DoorName,
		Delay: struct {
			Delay      Suffix `json:"delay"`
			Status     Suffix `json:"status"`
			Configured Suffix `json:"configured"`
			Error      Suffix `json:"error"`
			Modified   Suffix `json:"modified"`
		}{
			Delay:      DoorDelay,
			Status:     DoorDelayStatus,
			Configured: DoorDelayConfigured,
			Error:      DoorDelayError,
			Modified:   DoorDelayModified,
		},
		Control: struct {
			Control    Suffix `json:"control"`
			Status     Suffix `json:"status"`
			Configured Suffix `json:"configured"`
			Error      Suffix `json:"error"`
			Modified   Suffix `json:"modified"`
		}{
			Control:    DoorControl,
			Status:     DoorControlStatus,
			Configured: DoorControlConfigured,
			Error:      DoorControlError,
			Modified:   DoorControlModified,
		},
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
		OID:    EventsOID,
		Status: EventsStatus,
		First:  EventsFirst,
		Last:   EventsLast,

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

	Users: Users{
		OID: UsersOID,
		Metadata: Metadata{
			Status:   Status,
			Created:  Created,
			Deleted:  Deleted,
			Modified: Modified,
			Type:     Type,
		},
		Name:     UserName,
		UID:      UserUID,
		Role:     UserRole,
		Password: UserPassword,
		OTP:      UserOTP,
		OTPKey:   UserOTPKey,
		Locked:   UserLocked,
	},
}

const InterfacesOID OID = "0.1"
const ControllersOID OID = "0.2"
const DoorsOID OID = "0.3"
const CardsOID OID = "0.4"
const GroupsOID OID = "0.5"
const EventsOID OID = "0.6"
const LogsOID OID = "0.7"
const UsersOID OID = "0.8"

const Status Suffix = ".0.0"
const Created Suffix = ".0.1"
const Deleted Suffix = ".0.2"
const Modified Suffix = ".0.3"
const Type Suffix = ".0.4"
const Touched Suffix = ".0.5"

const InterfaceName Suffix = ".1"
const InterfaceID Suffix = ".2"
const LANBindAddress Suffix = ".3.1"
const LANBroadcastAddress Suffix = ".3.2"
const LANListenAddress Suffix = ".3.3"

const ControllerName Suffix = ".1"
const ControllerDeviceID Suffix = ".2"
const ControllerEndpoint Suffix = ".3"
const ControllerEndpointStatus Suffix = ".3.0"
const ControllerEndpointAddress Suffix = ".3.1"
const ControllerEndpointConfigured Suffix = ".3.2"
const ControllerDateTime Suffix = ".4" //TODO Fix when rationalizing the whole date/time/timezone mess
const ControllerDateTimeStatus Suffix = ".4.0"
const ControllerDateTimeCurrent Suffix = ".4.1"
const ControllerDateTimeConfigured Suffix = ".4.2"
const ControllerDateTimeModified Suffix = ".4.3"
const ControllerCardsStatus Suffix = ".5.0"
const ControllerCardsCount Suffix = ".5.1"
const ControllerEventsStatus Suffix = ".6.0"
const ControllerEventsFirst Suffix = ".6.1"
const ControllerEventsLast Suffix = ".6.2"
const ControllerEventsCurrent Suffix = ".6.3"
const ControllerDoor1 Suffix = ".7.1"
const ControllerDoor2 Suffix = ".7.2"
const ControllerDoor3 Suffix = ".7.3"
const ControllerDoor4 Suffix = ".7.4"

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

const EventsStatus Suffix = ".0.0"
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

const UserName Suffix = ".1"
const UserUID Suffix = ".2"
const UserRole Suffix = ".3"
const UserPassword Suffix = ".4"
const UserOTP Suffix = ".5"
const UserOTPKey Suffix = ".5.1"
const UserLocked Suffix = ".6"
