# `*.grl`

The _grules_ files supplement `auth.json` to provide fine grained authorisation for view, create, update and delete
operations. `auth.json` provides coarse-grained authorisation at the level of resource URLs e.g. a user can be allowed to view
either all of the cards on the system or none - nothing in between. the _grules_ file add fine-grained authorisation
at the _item_ level, which is useful when a user should only be able to view or modify a subset of the data e.g. a 
teacher may be restricted to viewing/editing only his/her/ze/their students.

The default _grules_ files comprise:

- `cards.grl`
- `controllers.grl`
- `doors.grl`
- `events.grl`
- `groups.grl`
- `interfaces.grl`
- `logs.grl`
- `users.grl`

and are embedded in the executable but can be overridden with external _grules_  files located (variously) in:

- /etc/uhppoted/httpd/grules (Linux)
- /usr/local/com.github.uhppoted/httpd/grules (MacOS)
- \Program Data\uhppoted\httpd\grules (Windows)

Rule-based access is currently implemented using the [Grule Rule Engine](https://github.com/hyperjumptech/grule-rule-engine),
the rules language documentation for which can be found [here](https://github.com/hyperjumptech/grule-rule-engine/blob/master/docs/Documentation.md).

A typical rule looks like:
```
rule ViewCard "(allowed)" {
     when
         OP == "view::card" && UID == 'McGonagall'
     then
         RESULT.Allow = true;
         Retract("ViewCard");
}
```

Permission for the operation will be granted if RESULT.Allow is true and RESULT.Refuse is not true - the default rulesets
grant permission for all items.

_grules_ files are automatically reloaded when modified i.e. it is not necessary to stop and restart `uhppoted-httpd`
for rule changes to take effect.

## Entities

The following _entities_ are maded available to the ACL _rules engine_:

- `OP`
- `UID` 
- `ROLE`
- `RESULT`
- `OBJECT`
- `FIELD`
- `VALUE`

### `OP` 

The `OP` entity is a readonly value of the form _operation:entity_ that defines the operation:

| operation | Description                                                            |
|-----------|------------------------------------------------------------------------|
| `view`    | Display item                                                           |
| `add`     | Create new item                                                        |
| `update`  | Modify existing item value                                             |
| `delete`  | Delete item                                                            |

| entity       | Description                                                         |
|--------------|---------------------------------------------------------------------|
| `interface`  | _interface_ that manages a set of controllers e.g. LAN              |
| `controller` | _controller_ attributes                                             |
| `door`       | _door_ configuration                                                |
| `card`       | _card_ details                                                      |
| `group`      | _group_ permissions                                                 |
| `user`       | _user_ attributes                                                   |
| `event`      | _controller events_                                                 |
| `log`        | _audit log_ records                                                 |

e.g. `update:card` would be the operation descriptor used when modifying a card name, number or other 
attribute.

### `UID`

The `UID` entity is simply the login user ID.

### `ROLE`

The `ROLE` entity is the role assigned to the login user ID (managed on the _users_ page). Common roles
include:

- admin
- user

but can be any valid string.

### `RESULT`

The `RESULT` entity contains the result of the evaluation of the ruleset and comprises two fields:

- `Allow`
- `Refuse`

Permission for the operation + entity is granted if `Allow` is true and `Refuse` is not true. The default
values are:

- `Allow`: true
- `Refuse`: false

### `OBJECT`

The `OBJECT` entity contains the identifying fields for the object for which the operation is being 
evaluated.

#### `interface`

| Field      | Description                                                            |
|------------|------------------------------------------------------------------------|
| `Type`     | _interface_ type e.g. LAN                                              |
| `Name`     | _interface_ name e.g. LANICA                                           |

#### `controller`

| Field      | Description                                                            |
|------------|------------------------------------------------------------------------|
| `Name`     | _controller_ name e.g. Alpha                                           |
| `DeviceID` | _controller_ serial number e.g. 405419896                              |

#### `door`

| Field      | Description                                                            |
|------------|------------------------------------------------------------------------|
| `Name`     | _door_ name e.g. Dungeon                                               |

#### `card`

| Field      | Description                                                            |
|------------|------------------------------------------------------------------------|
| `Name`     | _card holder_ name e.g. Dobby                                          |
| `Number`   | _card_ number e.g. 8165538                                             |
| `From`     | _card_ 'valid from' date as YYYY-MM-DD e.g. 2022-01-01                 |
| `To`       | _card_ 'valid until' date as YYYY-MM-DD e.g. 2022-12-31                |
| `Groups`   | _card_ groups membership list e.g. [ Student, Gryffindor ]             |

#### `group`

| Field      | Description                                                            |
|------------|------------------------------------------------------------------------|
| `Name`     | _group_ name e.g. Staff                                                |
| `Doors`    | Map of group doors permissions e.g. [Dungeon:false, Kitchen:true]      |

#### `user`

| Field      | Description                                                            |
|------------|------------------------------------------------------------------------|
| `Name`     | _user_ name e.g. Donald Duque                                          |
| `UID`      | _user_ login ID e.g. dduque                                            |
| `Role`     | _user_ role e.g. admin                                                 |

#### `event`

| Field      | Description                                                            |
|------------|------------------------------------------------------------------------|
| `Device`   | _event_ controller serial number e.g. 405419896                        |
| `Index`    | _event_ ID on controller e.g. 472                                      |

#### `log`

| Field       | Description                                                            |
|-------------|------------------------------------------------------------------------|
| `Timestamp` | _audit log record_ timestamp as YYYY-MM-DD HH:mm:ss ZZZ                |
|             | e.g. 2022-03-30 15:34:19 CET                                           |

### `FIELD`

The `FIELD` entity is the name of the objevt field for which the operation is being evaluated.

#### `interface`

| Field       | Description                                                           |
|-------------|-----------------------------------------------------------------------|
| `name`      | _interface_ name                                                      |
| `bind`      | _interface_ IPv4 UDP _bind_ address                                   |
| `broadcast` | _interface_ IPv4 UDP _broadcast_ address                              |
| `listen`    | _interface_ IPv4 UDP _listen_ address                                 |

#### `controller`

| Field       | Description                                                           |
|-------------|-----------------------------------------------------------------------|
| `name`      | _controller_ name                                                     |
| `deviceID`  | _controller_ serial number                                            |
| `address`   | _controller_ IPv4 address                                             |
| `timezone`  | _controller_ time zone                                                |
| `door[1]`   | _controller_ door 1 assignment                                        |
| `door[2]`   | _controller_ door 2 assignment                                        |
| `door[3]`   | _controller_ door 3 assignment                                        |
| `door[4]`   | _controller_ door 4 assignment                                        |

#### `door`

| Field       | Description                                                           |
|-------------|-----------------------------------------------------------------------|
| `name`      | _door_ name                                                           |
| `mode`      | _door_ control mode (normally open, normally closed or controlled)    |
| `delay`     | _door_ open delay                                                     |


#### `card`

| Field       | Description                                                           |
|-------------|-----------------------------------------------------------------------|
| `name`      | _card_ name                                                           |
| `number`    | _card_ number                                                         |
| `from`      | _card_ 'valid from' date (YYYY-MM-DD)                                 |
| `to`        | _card_ 'valid until' date (YYYY-MM-DD)                                |
| `group`     | _group_ name                                                          |

#### `group`

| Field       | Description                                                           |
|-------------|-----------------------------------------------------------------------|
| `name`      | _group_ name                                                          |
| _door_      | _door_ name e.g. Dungeon                                              |

#### `user`

| Field       | Description                                                           |
|-------------|-----------------------------------------------------------------------|
| `name`      | _user_ name                                                           |
| `uid`       | _user_ login ID                                                       |
| `password`  | _user_ login password                                                 |
| `role`      | _user_ role                                                           |

#### `event`

| Field        | Description                                                          |
|--------------|----------------------------------------------------------------------|
| `timestamp`  | _event_ timestamp                                                    |
| `type`       | _event_ type                                                         |
| `door`       | _event_ door ID                                                      |
| `direction`  | _event_ door direction                                               |
| `card`       | _event_ card number                                                  |
| `granted`    | _event_ access granted result                                        |
| `reason`     | _event_ reason                                                       |
| `deviceName` | _event_ controller name                                              |
| `doorName`   | _event_ door name                                                    |
| `cardName`   | _event_ card name                                                    |

#### `log`

| Field        | Description                                                          |
|--------------|----------------------------------------------------------------------|
| `timestamp`  | _audit record_ timestamp                                             |
| `UID`        | _audit record_ user login ID                                         |
| `item`       | _audit record_ entity OID                                            |
| `itemID`     | _audit record_ entity ID                                             |
| `itemName`   | _audit record_ entiy name                                            |
| `field`      | _audit record_ field name                                            |
| `details`    | _audit record_ operation description                                 |


### `VALUE`

The `VALUE` entity is only relevant for _update_ operations and contains the new value for the field.

## Sample `grules` file

```
rule ViewCard "(allowed)" {
     when
         OP == "view::card"
     then
         RESULT.Allow = true;
         Retract("ViewCard");
}

rule AddCard "(admin only)" {
     when
         OP == "add::card" && ROLE == 'admin'
     then
         RESULT.Allow = true;
         Retract("AddCard");
}

rule UpdateCard "(admin only)" {
     when
         OP == "update::card" && ROLE == 'admin'
     then
         RESULT.Allow = true;
         Retract("UpdateCard");
}

rule DeleteCard "(admin)" {
     when
         OP == "delete::card" && ROLE == 'admin'
     then
         RESULT.Allow = true;
         Retract("DeleteCard");
}
```