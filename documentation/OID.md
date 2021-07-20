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

0                                                              # 'root'
|
|- 0.1                                                         # 'system'
     |
     |- 0.1.1                                                  # controller set
     |      |
     |      |- 0.1.1.0: 'LAN'                                  # controller set interface
     |      |        |- 0.1.1.0.1: <name>                      # interface name
     |      |        |- 0.1.1.1.2: <bind>                      # bind address
     |      |        |- 0.1.1.1.3: <broadcast>                 # broadcast address
     |      |        |- 0.1.1.1.4: <listen>                    # listen address
     |      |
     |      |- 0.1.1.1: <status>                               # UHPPOTE controller #1
     |      |        |- 0.1.1.1.0                              #
     |      |        |          |- 0.1.1.1.0.1: <created>      # created date/time
     |      |        |                                         #
     |      |        |- 0.1.1.1.1:  <name>                     # name
     |      |        |- 0.1.1.1.2:  <ID>                       # serial number
     |      |        |- 0.1.1.1.3:  <address>                  # address:port
     |      |                   |- 0.1.1.1.3.1:  <configured>  # configured address:port
     |      |        |- 0.1.1.1.4:  <datetime>                 # system date/time
     |      |        |- 0.1.1.1.5:  <cards>                    # number of card records
     |      |        |- 0.1.1.1.6:  <events>                   # number of event records
     |      |        |- 0.1.1.1.7:  <door1>                    # door 1 name
     |      |        |- 0.1.1.1.8:  <door2>                    # door 2 name
     |      |        |- 0.1.1.1.9:  <door3>                    # door 3 name
     |      |        |- 0.1.1.1.10: <door4>                    # door 4 name
     |      |
     |      |- 0.1.1.2: <status>                               # UHPPOTE controller #2
     |      |        |- 0.1.1.2.0                              #
     |      |        |          |- 0.1.1.2.0.1: <created>      # created date/time
     |      |        |                                         #
     |      |        |- 0.1.1.2.1:  <name>                     # name
     |      |        |- 0.1.1.2.2:  <ID>                       # serial number
     |      |        |- 0.1.1.2.3:  <address>                  # address:port
     |      |                   |- 0.1.1.2.3.1:  <configured>  # configured address:port
     |      |        |- 0.1.1.2.4:  <datetime>                 # system date/time
     |      |        |- 0.1.1.2.5:  <cards>                    # number of card records
     |      |        |- 0.1.1.2.6:  <events>                   # number of event records
     |      |        |- 0.1.1.2.7:  <door1>                    # door 1 name
     |      |        |- 0.1.1.2.8:  <door2>                    # door 2 name
     |      |        |- 0.1.1.2.9:  <door3>                    # door 3 name
     |      |        |- 0.1.1.2.10: <door4>                    # door 4 name
     |      |
     |      |- ...
     |
     |- 0.1.2                                                  # controller set
     |      |
     |
     |- ...
     | 





