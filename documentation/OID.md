# OID Namespace
```
0                                                                            # root
|
|- 0.0                                                                       # system
|    |- 0.1:                                                                 #   cards
|         |- 0.1.1: _default start date_                                     #     default start date
|         |- 0.1.2: _default end date_                                       #     default end date
|
|- 0.1                                                                       # interfaces
|    |- 0.1.1:                                                               # interface #1
|    |      |- 0.1.1.0: _metadata_                                           #    metadata
|    |      |        |- 0.1.1.0.0: _status_                                  #       current status
|    |      |        |- 0.1.1.0.1: _created_                                 #       created date/time
|    |      |        |- 0.1.1.0.2: _deleted_                                 #       deleted date/time
|    |      |        |- 0.1.1.0.3: _modified_                                #       modified timestamp
|    |      |        |- 0.1.1.0.4: _type_                                            type
|    |      |                                                                #
|    |      |- 0.1.1.1: _name_                                               #    name
|    |      |- 0.1.1.2: _ID_                                                 #    ID
|    |      |- 0.1.1.3: _LAN_                                                #
|    |      |        |- 0.1.1.3.1: _bind_                                    #    LAN bind address
|    |      |        |- 0.1.1.3.2: _broadcast_                               #    LAN broadcast address
|    |      |        |- 0.1.1.3.3: _listen_                                  #    LAN listen address
|    |- ...
| 
|- 0.2                                                                       # boards
|    |- 0.2.1:                                                               # board #1
|    |      |- 0.2.1.0: _metadata_                                           #    metadata
|    |      |        |- 0.2.1.0.0: _status_                                  #       current status
|    |      |        |- 0.2.1.0.1: _created_                                 #       created date/time
|    |      |        |- 0.2.1.0.2: _deleted_                                 #       deleted date/time
|    |      |        |- 0.2.1.0.3: _modified_                                #       modified timestamp
|    |      |        |- 0.2.1.0.4: _type_                                    #    type
|    |      |        |- 0.2.1.0.5: _touched_                                 #    refreshed timestamp
|    |      |
|    |      |- 0.2.1.1: _name_                                               #    name
|    |      |- 0.2.1.2: _ID_                                                 #    serial number
|    |      |- 0.2.1.3: _address_                                            #    controller endpoint
|    |      |        |- 0.2.1.3.0: _status_                                  #       address status
|    |      |        |- 0.2.1.3.1: _endpoint_                                #       controller address:port
|    |      |        |- 0.2.1.3.2: _configured_                              #       configured address:port
|    |      |        |- 0.2.1.3.3: _protocol_                                #       transport (TCP or UDP)
|    |      |- 0.2.1.4:  _datetime_                                          #    controller date/time
|    |      |        |- 0.2.1.4.0: _status_                                  #       status
|    |      |        |- 0.2.1.4.1: _current_                                 #       controller date/time
|    |      |        |- 0.2.1.5.2: _system_                                  #       system date/time
|    |      |        |- 0.2.1.6.3: _modified_                                #       modified
|    |      |- 0.2.1.5:  _cards_                                             #    controller cards
|    |      |        |- 0.2.1.5.0: _status_                                  #       cards status
|    |      |        |- 0.2.1.5.1: _count_                                   #       number of card
|    |      |- 0.2.1.6:  _events_                                            #    controller events
|    |      |        |- 0.2.1.6.0: _status_                                  #       events status
|    |      |        |- 0.2.1.6.1: _first_                                   #       index of first event
|    |      |        |- 0.2.1.6.2: _last_                                    #       index of last event
|    |      |        |- 0.2.1.6.3: _current_                                 #       index of current event
|    |      |- 0.2.1.7:  _doors_                                             #    doors
|    |      |        |- 0.2.1.7.1: _door1_                                   #       door 1 OID
|    |      |        |- 0.2.1.7.2: _door2_                                   #       door 2 OID
|    |      |        |- 0.2.1.7.3: _door3_                                   #       door 3 OID
|    |      |        |- 0.2.1.7.4: _door4_                                   #       door 4 OID
|    |      |- 0.2.1.8:  _interlock_                                         #    doors interlock mode
|    |      |- 0.2.1.9:  _antipassback_                                      #    card anti-passback mode
|    |      |        |- 0.2.1.9.0: _status_                                  #       anti-passback status (ok, uncertain, error)
|    |      |        |- 0.2.1.9.1: _antipassback_                            #       current anti-passback mode
|    |      |        |- 0.2.1.9.2: _configured_                              #       configured anti-passback mode
|    |      |        |- 0.2.1.9.3: _modified_                                #       anti-passback modified
|    |- ...
|
|- 0.3                                                                       # doors
|    |- 0.3.1:                                                               # door #1
|    |      |- 0.3.1.0: _metadata_                                           #    metadata
|    |      |        |- 0.3.1.0.0: _status_                                  #       current status
|    |      |        |- 0.3.1.0.1: _created_                                 #       created date/time
|    |      |        |- 0.3.1.0.2: _deleted_                                 #       deleted date/time
|    |      |        |- 0.3.1.0.3: _modified_                                #       modified timestamp
|    |      |        |- 0.3.1.0.4: _controller_                              #       associated controller OID
|    |      |                   |- 0.3.1.0.4.1: _created_                    #               controller created date/time
|    |      |                   |- 0.3.1.0.4.2: _name_                       #               controller name
|    |      |                   |- 0.3.1.0.4.3: _deviceID_                   #               controller serial number
|    |      |                   |- 0.3.1.0.4.4: _door_                       #               controller door number
|    |      |                                                                #
|    |      |- 0.3.1.1: _name_                                               #    name
|    |      |- 0.3.1.2: _delay_                                              #    door open delay value
|    |               |- 0.3.1.2.1: _status_                                  #                    status
|    |               |- 0.3.1.2.2: _configured_                              #                    configured value
|    |               |- 0.3.1.2.3: _error_                                   #                    error info
|    |               |- 0.3.1.2.4: _modified_                                #                    has been modified
|    |      |- 0.3.1.3: _control_                                            #    door control state value
|    |               |- 0.3.1.2.1: _status_                                  #    door control state status
|    |               |- 0.3.1.2.2: _configured_                              #                       configured value
|    |               |- 0.3.1.2.3: _error_                                   #                       error info
|    |               |- 0.3.1.2.4: _modified_                                #                       has been modified
|
|- 0.4                                                                       # cards
|    |- 0.4.1:                                                               # card #1
|    |      |- 0.4.1.0: _metadata_                                           #    metadata
|    |      |        |- 0.4.1.0.0: _status_                                  #       current status
|    |      |        |- 0.4.1.0.1: _created_                                 #       created timestamp
|    |      |        |- 0.4.1.0.2: _deleted_                                 #       deleted timestamp
|    |      |        |- 0.4.1.0.3: _modified_                                #       modified timestamp
|    |      |                                                                # 
|    |      |- 0.4.1.1: _name_                                               #      name
|    |      |- 0.4.1.2: _number_                                             #      card number
|    |      |- 0.4.1.3: _from_                                               #      'valid from' date
|    |      |- 0.4.1.4: _to_                                                 #      'valid until' date
|    |      |- 0.4.1.5                                                       #      groups
|    |               |- 0.4.1.5.1 _member_                                   #      group #1: member
|    |               |           |- 0.4.1.5.1.1: _oid_                       #                group OID
|    |               |                                                       #
|    |               |- ...                                                  #      group #2...
|    |- ...
|
|- 0.5                                                                       # groups
|    |- 0.5.1                                                                # group #1
|    |      |- 0.5.1.0: _metadata_                                           #    metadata
|    |      |        |- 0.5.1.0.0: _status_                                  #       current status
|    |      |        |- 0.5.1.0.1: _created_                                 #       created date/time
|    |      |        |- 0.5.1.0.2: _deleted_                                 #       deleted date/time
|    |      |        |- 0.5.1.0.3: _modified_                                #       modified timestamp
|    |      |- 0.5.1.1: _name_                                               #       Name
|    |      |- 0.5.1.2: _index_                                              #       Index (display order)
|    |- ...
|
|- 0.6                                                                       # events
|    |- 0.6.0                                                                #    metadata
|    |      |- 0.6.0.0: _status_                                             #       synthesized status
|    |      |- 0.6.0.1                                                       #       first event OID
|    |      |- 0.6.0.2                                                       #       last event OID
|    |
|    |- 0.6.1                                                                #    event #1
|    |      |- 0.6.1.1:  _timestamp_                                         #       event timestamp
|    |      |- 0.6.1.2:  _deviceID_                                          #       device ID
|    |      |- 0.6.1.3:  _index_                                             #       event index
|    |      |- 0.6.1.4:  _type_                                              #       event type
|    |      |- 0.6.1.5:  _door_                                              #       event door ID
|    |      |- 0.6.1.6:  _direction_                                         #       event direction
|    |      |- 0.6.1.7:  _cardNumber_                                        #       event card Number
|    |      |- 0.6.1.8:  _accessGranted_                                     #       event access granted
|    |      |- 0.6.1.9:  _reason_                                            #       event access reason
|    |      |- 0.6.1.10: _deviceName_                                        #       (associated) controller name
|    |      |- 0.6.1.11: _doorName_                                          #       (associated) door name
|    |      |- 0.6.1.12: _cardName_                                          #       (associated) card holder
|    |- ...
|
|- 0.7                                                                       # logs
|    |- 0.7.0                                                                #    metadata
|    |      |- 0.7.0.1                                                       #       first log entry OID
|    |      |- 0.7.0.2                                                       #       last log entry OID
|    |
|    |- 0.7.1                                                                #    entry #1
|    |      |- 0.7.1.1: _timestamp_                                          #       timestamp
|    |      |- 0.7.1.2: _uid_                                                #       user ID
|    |      |- 0.7.1.3: _item_                                               #       item type
|    |      |- 0.7.1.4: _id_                                                 #       item ID
|    |      |- 0.7.1.5: _name_                                               #       item name
|    |      |- 0.7.1.6: _field_                                              #       item field
|    |      |- 0.7.1.7: _details_                                            #       item details
|
|- 0.8                                                                       # users
|    |- 0.8.1                                                                # user #1
|    |      |- 0.8.1.0: _metadata_                                           #    metadata
|    |      |        |- 0.8.1.0.0: _status_                                  #       status
|    |      |        |- 0.8.1.0.1: _created_                                 #       created date/time
|    |      |        |- 0.8.1.0.2: _deleted_                                 #       deleted date/time
|    |      |        |- 0.8.1.0.3: _modified_                                #       modified timestamp
|    |      |- 0.8.1.1: _name_                                               #       Name
|    |      |- 0.8.1.2: _uid_                                                #       UID
|    |      |- 0.8.1.3: _role_                                               #       Role
|    |      |- 0.8.1.4: _password_                                           #       Password
|    |      |- 0.8.1.5: _otp_                                                #       OTP (enabled)
|    |      |        |- 0.8.1.5.1: _otpKey_                                  #       OTP secret key
|    |      |- 0.8.1.6: _locked_                                             #       locked flag
|    |- ...
|

```