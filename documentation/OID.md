# Object ID Namespace

Hierarchical ID structure modelled on the OID structure of SNMP:

## 0.x.x.x

Global namespace

### 0.1.x.x

System namespace

#### 0.1.1.x

ControllerSet namespace

#### 0.1.1.1.x

ControllerSet interface namespace

#### 0.1.1.2.x

Controller namespace

### 0.2.x.x

Doors namespace

### 0.3.x.x

Cards namespace

### 0.4.x.x

Groups namespace

### 0.5.x.x

Events namespace

### 0.6.x.x

Logs namespace

### 0.7.x.x

# OID tree

0                                                                            # root
|
|- 0.1                                                                       # system
|    |
|    |- 0.1.1                                                                # controller set
|    |      |
|    |      |- 0.1.1.1:                                                      # interfaces
|    |      |        |
|    |      |        |- 0.1.1.1.1: <status>                                  # interface #1
|    |      |                   |- 0.1.1.1.1.0: <type>                       #    type
|    |      |                   |- 0.1.1.1.1.1: <name>                       #    name
|    |      |                   |- 0.1.1.1.1.2: <bind>                       #    bind address
|    |      |                   |- 0.1.1.1.1.3: <broadcast>                  #    broadcast address
|    |      |                   |- 0.1.1.1.1.4: <listen>                     #    listen address
|    |      |
|    |      |- 0.1.1.2:                                                      # boards
|    |               |
|    |               |- 0.1.1.2.1: <status>                                  # board #1
|    |               |          |- 0.1.1.2.1.0: <type>                       #    type
|    |               |          |            |- 0.1.1.2.1.0.1: <created>     #    created date/time
|    |               |          |                                            #
|    |               |          |- 0.1.1.2.1.1:  <name>                      #    name
|    |               |          |- 0.1.1.2.1.2:  <ID>                        #    serial number
|    |               |          |- 0.1.1.2.1.3:  <address>                   #    address:port
|    |               |                       |- 0.1.1.1.2.3.1: <configured>  #    configured address:port
|    |               |                       |- 0.1.1.1.2.3.2: <status>      #    address status
|    |               |          |- 0.1.1.2.1.4:  <datetime>                  #    controller date/time
|    |               |                       |- 0.1.1.1.2.4.1: <now>         #    system date/time
|    |               |                       |- 0.1.1.1.2.4.2: <status>      #    controller date/time status
|    |               |          |- 0.1.1.2.1.5:  <cards>                     #    number of card records
|    |               |                       |- 0.1.1.1.2.5.1: <status>      #    cards status
|    |               |          |- 0.1.1.2.1.6:  <events>                    #    number of event records
|    |               |                       |- 0.1.1.1.2.6.1: <status>      #    events status
|    |               |          |- 0.1.1.2.1.7:  <door1>                     #    door 1 (OID)
|    |               |          |- 0.1.1.2.1.8:  <door2>                     #    door 2 (OID)
|    |               |          |- 0.1.1.2.1.9:  <door3>                     #    door 3 (OID)
|    |               |          |- 0.1.1.2.1.10: <door4>                     #    door 4 (OID)
|    |               |
|    |               |- 0.1.1.2.2: <status>                                  # board #2
|    |               |          |- ...
|    |               |
|    |               |- ...
|    |
|    |- ...
|
|- 0.2                                                                       # doors
|    |- 0.2.1: <status>                                                      # door #1
|    |      |- 0.2.1.0:                                                      #
|    |      |        |- 0.2.1.0.1: <created>                                 #    created date/time
|    |      |        |- 0.2.1.0.2: <controller>                              #    associated controller OID
|    |      |                   |- 0.2.1.0.2.1: <created>                    #               controller created date/time
|    |      |                   |- 0.2.1.0.2.2: <name>                       #               controller name
|    |      |                   |- 0.2.1.0.2.3: <deviceID>                   #               controller serial number
|    |      |                   |- 0.2.1.0.2.4: <door>                       #               controller door number
|    |      |                                                                #
|    |      |- 0.2.1.1: <name>                                               #    name
|    |      |- 0.2.1.2: <delay>                                              #    door open delay
|    |               |- 0.2.1.2.1: <status>                                  #    door open delay status
|    |               |- 0.2.1.2.2: <configured>                              #    configured door open delay
|    |               |- 0.2.1.2.3: <error>                                   #    door delay error info
|    |      |- 0.2.1.3: <control>                                            #    door control state
|    |               |- 0.2.1.2.1: <status>                                  #    door control state status
|    |               |- 0.2.1.2.2: <configured>                              #    configured door control state
|    |               |- 0.2.1.2.3: <error>                                   #    door control state error info
|
|- 0.3                                                                       # cards
|    |- 0.3.1: <status>                                                      # card #1
|    |      |- 0.3.1.0:                                                      #
|    |      |        |- 0.3.1.0.1: <created>                                 #      created date/time
|    |      |                                                                # 
|    |      |- 0.3.1.1: <name>                                               #      name
|    |      |- 0.3.1.2: <number>                                             #      card number
|    |      |- 0.3.1.3: <from>                                               #      'valid from' date
|    |      |- 0.3.1.4: <to>                                                 #      'valid until' date
|    |      |- 0.3.1.5                                                       #      groups
|    |               |- 0.3.1.5.1 <member>                                   #      group #1: member
|    |               |           |- 0.3.1.5.1.1: <oid>                       #                group OID
|    |               |                                                       #
|    |               |- ...                                                  #      group #2...
|    |- ...
|
|- 0.4                                                                       # groups
|    |- 0.4.1                                                                # group #1
|    |      |- 0.4.1.1: <name>                                               #       Name
|    |      |- 0.4.1.2: <index>                                              #       Index (display order)
|
|- 0.5                                                                       # events
|    |- 0.5.0                                                                # 
|    |      |- 0.5.0.1                                                       # first event OID
|    |      |- 0.5.0.2                                                       # last event OID
|    |
|    |- 0.5.1                                                                # event #1
|    |      |- 0.5.1.1: <timestamp>                                          #       event timestamp
|    |      |- 0.5.1.2: <deviceID>                                           #       device ID
|    |      |- 0.5.1.3: <index>                                              #       event index
|    |      |- 0.5.1.4: <type>                                               #       event type
|    |      |- 0.5.1.5: <door>                                               #       event door ID
|    |      |- 0.5.1.6: <direction>                                          #       event direction
|    |      |- 0.5.1.7: <cardNumber>                                         #       event card Number
|    |      |- 0.5.1.8: <accessGranted>                                      #       event access granted
|    |      |- 0.5.1.9: <reason>                                             #       event access reason
|    |      |- 0.5.1.10: <deviceName>                                        #       (associated) controller name
|    |      |- 0.5.1.11: <doorName>                                          #       (associated) door name
|    |      |- 0.5.1.12: <cardName>                                          #       (associated) card holder
|
|- 0.6                                                                       # logs
|    |- 0.6.0                                                                # 
|    |      |- 0.6.0.1                                                       # first log entry OID
|    |      |- 0.6.0.2                                                       # last log entry OID
|    |
|    |- 0.6.1                                                                # entry #1
|    |      |- 0.6.1.1: <timestamp>                                          #       entry timestamp



