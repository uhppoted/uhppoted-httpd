# Object ID Namespace

Hierarchical ID structure modelled on the OID structure of SNMP:

## 0.x.x.x

Global namespace

### 0.1.x.x

System namespace

#### 0.1.1.x

ControllerSet namespace

#### 0.1.1.0.x

ControllerSet interface namespace

#### 0.1.1.1.x

Controller namespace

### 0.2.x.x

Cards namespace

### 0.3.x.x

Doors/access namespace

### 0.4.x.x

Events namespace

### 0.5.x.x

Logs namespace

# OID tree

0                                                           # 'root'
|
|- 0.1                                                      # 'system'
     |
     |- 0.1.1                                               # controller set
     |      |
     |      |- 0.1.1.0: 'LAN'                               # controller set interface
     |      |        |- 0.1.1.0.1: <name>                   # interface name
     |      |        |- 0.1.1.1.2: <bind>                   # bind address
     |      |        |- 0.1.1.1.3: <broadcast>              # broadcast address
     |      |        |- 0.1.1.1.4: <listen>                 # listen address
     |      |
     |      |- 0.1.1.1: <status>                            # UHPPOTE controller
     |      |        |- 0.1.1.1.0                           #
     |      |        |          |- 0.1.1.1.0.1: <created>   # created date/time
     |      |        |                                      #
     |      |        |- 0.1.1.1.1:  <name>                  # UHPPOTE controller name
     |      |        |- 0.1.1.1.2:  <ID>                    # UHPPOTE controller serial number
     |      |        |- 0.1.1.1.3:  <address>               # UHPPOTE controller address:port
     |      |        |- 0.1.1.1.4:  <datetime>              # UHPPOTE controller system date/time
     |      |        |- 0.1.1.1.5:  <cards>                 # UHPPOTE controller number of card records
     |      |        |- 0.1.1.1.6:  <events>                # UHPPOTE controller number of event records
     |      |        |- 0.1.1.1.7:  <door1>                 # UHPPOTE controller door 1 name
     |      |        |- 0.1.1.1.8:  <door2>                 # UHPPOTE controller door 2 name
     |      |        |- 0.1.1.1.9:  <door3>                 # UHPPOTE controller door 3 name
     |      |        |- 0.1.1.1.10: <door4>                 # UHPPOTE controller door 4 name
     |      |
     |      |- 0.1.1.2: <status>                            # UHPPOTE controller
     |      |        |- 0.1.1.2.0                           #
     |      |        |          |- 0.1.1.2.0.1: <created>   # created date/time
     |      |        |                                      #
     |      |        |- 0.1.1.2.1:  <name>                  # UHPPOTE controller name
     |      |        |- 0.1.1.2.2:  <ID>                    # UHPPOTE controller serial number
     |      |        |- 0.1.1.2.3:  <address>               # UHPPOTE controller address:port
     |      |        |- 0.1.1.2.4:  <datetime>              # UHPPOTE controller system date/time
     |      |        |- 0.1.1.2.5:  <cards>                 # UHPPOTE controller number of card records
     |      |        |- 0.1.1.2.6:  <events>                # UHPPOTE controller number of event records
     |      |        |- 0.1.1.2.7:  <door1>                 # UHPPOTE controller door 1 name
     |      |        |- 0.1.1.2.8:  <door2>                 # UHPPOTE controller door 2 name
     |      |        |- 0.1.1.2.9:  <door3>                 # UHPPOTE controller door 3 name
     |      |        |- 0.1.1.2.10: <door4>                 # UHPPOTE controller door 4 name2
     |      |
     |      |- ...
     |
     |- 0.1.2                                               # controller set
     |      |
     |
     |- ...
     | 





