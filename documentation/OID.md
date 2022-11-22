# OID Namespace
```
0                                                                            # root
|
|- 0.1                                                                       # interfaces
|    |- 0.1.1:                                                               # interface #1
|    |      |- 0.1.1.0: <metadata>                                           #    metadata
|    |      |        |- 0.1.1.0.0: <status>                                  #       current status
|    |      |        |- 0.1.1.0.1: <created>                                 #       created date/time
|    |      |        |- 0.1.1.0.2: <deleted>                                 #       deleted date/time
|    |      |        |- 0.1.1.0.3: <modified>                                #       modified timestamp
|    |      |        |- 0.1.1.0.4: <type>                                            type
|    |      |                                                                #
|    |      |- 0.1.1.1: <name>                                               #    name
|    |      |- 0.1.1.2: <ID>                                                 #    ID
|    |      |- 0.1.1.3: <LAN>                                                #
|    |      |        |- 0.1.1.3.1: <bind>                                    #    LAN bind address
|    |      |        |- 0.1.1.3.2: <broadcast>                               #    LAN broadcast address
|    |      |        |- 0.1.1.3.3: <listen>                                  #    LAN listen address
|    |- ...
| 
|- 0.2                                                                       # boards
|    |- 0.2.1:                                                               # board #1
|    |      |- 0.2.1.0: <metadata>                                           #    metadata
|    |      |        |- 0.2.1.0.0: <status>                                  #       current status
|    |      |        |- 0.2.1.0.1: <created>                                 #       created date/time
|    |      |        |- 0.2.1.0.2: <deleted>                                 #       deleted date/time
|    |      |        |- 0.2.1.0.3: <modified>                                #       modified timestamp
|    |      |        |- 0.2.1.0.4: <type>                                    #    type
|    |      |        |- 0.2.1.0.5: <touched>                                 #    refreshed timestamp
|    |      |
|    |      |- 0.2.1.1: <name>                                               #    name
|    |      |- 0.2.1.2: <ID>                                                 #    serial number
|    |      |- 0.2.1.3: <address>                                            #    controller endpoint
|    |      |        |- 0.2.1.4.0: <status>                                  #       address status
|    |      |        |- 0.2.1.4.1: <endpoint>                                #       controller address:port
|    |      |        |- 0.2.1.4.2: <configured>                              #       configured address:port
|    |      |- 0.2.1.4:  <datetime>                                          #    controller date/time
|    |      |        |- 0.2.1.5.0: <status>                                  #       status
|    |      |        |- 0.2.1.5.1: <current>                                 #       controller date/time
|    |      |        |- 0.2.1.5.2: <system>                                  #       system date/time
|    |      |        |- 0.2.1.5.3: <modified>                                #       modified
|    |      |- 0.2.1.5:  <cards>                                             #    controller cards
|    |      |        |- 0.2.1.5.0: <status>                                  #       cards status
|    |      |        |- 0.2.1.5.1: <count>                                   #       number of card
|    |      |- 0.2.1.6:  <events>                                            #    controller events
|    |      |        |- 0.2.1.6.0: <status>                                  #       events status
|    |      |        |- 0.2.1.6.1: <first>                                   #       index of first event
|    |      |        |- 0.2.1.6.2: <last>                                    #       index of last event
|    |      |        |- 0.2.1.6.3: <current>                                 #       index of current event
|    |      |- 0.2.1.7:  <doors>                                             #    doors
|    |      |        |- 0.2.1.7.1: <door1>                                   #       door 1 OID
|    |      |        |- 0.2.1.7.2: <door2>                                   #       door 2 OID
|    |      |        |- 0.2.1.7.3: <door3>                                   #       door 3 OID
|    |      |        |- 0.2.1.7.4: <door4>                                   #       door 4 OID
|    |- ...
|
|- 0.3                                                                       # doors
|    |- 0.3.1:                                                               # door #1
|    |      |- 0.3.1.0: <metadata>                                           #    metadata
|    |      |        |- 0.3.1.0.0: <status>                                  #       current status
|    |      |        |- 0.3.1.0.1: <created>                                 #       created date/time
|    |      |        |- 0.3.1.0.2: <deleted>                                 #       deleted date/time
|    |      |        |- 0.3.1.0.3: <modified>                                #       modified timestamp
|    |      |        |- 0.3.1.0.4: <controller>                              #       associated controller OID
|    |      |                   |- 0.3.1.0.4.1: <created>                    #               controller created date/time
|    |      |                   |- 0.3.1.0.4.2: <name>                       #               controller name
|    |      |                   |- 0.3.1.0.4.3: <deviceID>                   #               controller serial number
|    |      |                   |- 0.3.1.0.4.4: <door>                       #               controller door number
|    |      |                                                                #
|    |      |- 0.3.1.1: <name>                                               #    name
|    |      |- 0.3.1.2: <delay>                                              #    door open delay value
|    |               |- 0.3.1.2.1: <status>                                  #                    status
|    |               |- 0.3.1.2.2: <configured>                              #                    configured value
|    |               |- 0.3.1.2.3: <error>                                   #                    error info
|    |               |- 0.3.1.2.4: <modified>                                #                    has been modified
|    |      |- 0.3.1.3: <control>                                            #    door control state value
|    |               |- 0.3.1.2.1: <status>                                  #    door control state status
|    |               |- 0.3.1.2.2: <configured>                              #                       configured value
|    |               |- 0.3.1.2.3: <error>                                   #                       error info
|    |               |- 0.3.1.2.4: <modified>                                #                       has been modified
|
|- 0.4                                                                       # cards
|    |- 0.4.1:                                                               # card #1
|    |      |- 0.4.1.0: <metadata>                                           #    metadata
|    |      |        |- 0.4.1.0.0: <status>                                  #       current status
|    |      |        |- 0.4.1.0.1: <created>                                 #       created timestamp
|    |      |        |- 0.4.1.0.2: <deleted>                                 #       deleted timestamp
|    |      |        |- 0.4.1.0.3: <modified>                                #       modified timestamp
|    |      |                                                                # 
|    |      |- 0.4.1.1: <name>                                               #      name
|    |      |- 0.4.1.2: <number>                                             #      card number
|    |      |- 0.4.1.3: <from>                                               #      'valid from' date
|    |      |- 0.4.1.4: <to>                                                 #      'valid until' date
|    |      |- 0.4.1.5                                                       #      groups
|    |               |- 0.4.1.5.1 <member>                                   #      group #1: member
|    |               |           |- 0.4.1.5.1.1: <oid>                       #                group OID
|    |               |                                                       #
|    |               |- ...                                                  #      group #2...
|    |- ...
|
|- 0.5                                                                       # groups
|    |- 0.5.1                                                                # group #1
|    |      |- 0.5.1.0: <metadata>                                           #    metadata
|    |      |        |- 0.5.1.0.0: <status>                                  #       current status
|    |      |        |- 0.5.1.0.1: <created>                                 #       created date/time
|    |      |        |- 0.5.1.0.2: <deleted>                                 #       deleted date/time
|    |      |        |- 0.5.1.0.3: <modified>                                #       modified timestamp
|    |      |- 0.5.1.1: <name>                                               #       Name
|    |      |- 0.5.1.2: <index>                                              #       Index (display order)
|    |- ...
|
|- 0.6                                                                       # events
|    |- 0.6.0                                                                #    metadata
|    |      |- 0.6.0.0: <status>                                             #       synthesized status
|    |      |- 0.6.0.1                                                       #       first event OID
|    |      |- 0.6.0.2                                                       #       last event OID
|    |
|    |- 0.6.1                                                                #    event #1
|    |      |- 0.6.1.1:  <timestamp>                                         #       event timestamp
|    |      |- 0.6.1.2:  <deviceID>                                          #       device ID
|    |      |- 0.6.1.3:  <index>                                             #       event index
|    |      |- 0.6.1.4:  <type>                                              #       event type
|    |      |- 0.6.1.5:  <door>                                              #       event door ID
|    |      |- 0.6.1.6:  <direction>                                         #       event direction
|    |      |- 0.6.1.7:  <cardNumber>                                        #       event card Number
|    |      |- 0.6.1.8:  <accessGranted>                                     #       event access granted
|    |      |- 0.6.1.9:  <reason>                                            #       event access reason
|    |      |- 0.6.1.10: <deviceName>                                        #       (associated) controller name
|    |      |- 0.6.1.11: <doorName>                                          #       (associated) door name
|    |      |- 0.6.1.12: <cardName>                                          #       (associated) card holder
|    |- ...
|
|- 0.7                                                                       # logs
|    |- 0.7.0                                                                #    metadata
|    |      |- 0.7.0.1                                                       #       first log entry OID
|    |      |- 0.7.0.2                                                       #       last log entry OID
|    |
|    |- 0.7.1                                                                #    entry #1
|    |      |- 0.7.1.1: <timestamp>                                          #       timestamp
|    |      |- 0.7.1.2: <uid>                                                #       user ID
|    |      |- 0.7.1.3: <item>                                               #       item type
|    |      |- 0.7.1.4: <id>                                                 #       item ID
|    |      |- 0.7.1.5: <name>                                               #       item name
|    |      |- 0.7.1.6: <field>                                              #       item field
|    |      |- 0.7.1.7: <details>                                            #       item details
|
|- 0.8                                                                       # users
|    |- 0.8.1                                                                # user #1
|    |      |- 0.8.1.0: <metadata>                                           #    metadata
|    |      |        |- 0.8.1.0.0: <status>                                  #       status
|    |      |        |- 0.8.1.0.1: <created>                                 #       created date/time
|    |      |        |- 0.8.1.0.2: <deleted>                                 #       deleted date/time
|    |      |        |- 0.8.1.0.3: <modified>                                #       modified timestamp
|    |      |- 0.8.1.1: <name>                                               #       Name
|    |      |- 0.8.1.2: <uid>                                                #       UID
|    |      |- 0.8.1.3: <role>                                               #       Role
|    |      |- 0.8.1.4: <password>                                           #       Password
|    |      |- 0.8.1.5: <otp>                                                #       OTP (enabled)
|    |               |- 0.8.1.5.1: <otpKey>                                  #       OTP secret key
|    |- ...
|

```