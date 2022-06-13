# `acl.grule`

The `acl.grule` adds rule-base access control to supplement the relatively simple 'grid-based' access control provided by
the groups + doors combination of the user interface. It is useful for applications that require e.g.:

- special access for selected cards
- access rules that are difficult to express as group privileges

Rule-based access is currently implemented using the [Grule Rule Engine](https://github.com/hyperjumptech/grule-rule-engine),
the rules language documentation for which can be found [here](https://github.com/hyperjumptech/grule-rule-engine/blob/master/docs/Documentation.md).

The access control permissions for a card that are set by the card's group membership may be overriden by one or more rules
that either:

- `Allow` access to a door e.g.
```
         DOORS.Allow("Dungeon");
```

- `Revoke` access to a door
```
         DOORS.Revoke("Hogsmeade");
```

Access to a door will be granted to a card if:

- one of the groups assigned to the card has permission for the door and the access is not _revoked_ in the `acl.grules` file
- the card has access to the door _allowed_ by at least one rule and also not _revoked_ by any rule

The `acl.grules` file is automatically reloaded when it has been modified i.e. it is not necessary to stop and restart `uhppoted-httpd`
for rule changes to take effect.

## Objects

The following _entities_ are maded available to the ACL _rules engine_:

- `Card`
- `Doors`
- `Query`

### `Card` 

The `Card` entity supplied to the _rules engined_ has the following fields:

| Field    | Description                                                         |
|----------|---------------------------------------------------------------------|
| `Name`   | Name of card holder                                                 |
| `Number` | Card number                                                         |
| `From`   | Date from which the card is valid (YYYY-MM-DD, inclusive)           |
| `To`     | Date after which the card is no longe valid (YYYY-DD_MM, inclusive) |
| `Groups` | Access control groups assigned to the card                          |


### `Doors`

The `Doors` entity supplied to the _rules engined_ comprises two arrays:
- `allowed`
- `forbidden`

A card can be given access rights to a door by adding it to the `allowed` list using
``` 
Doors.Allow(<door name>)
```

A card can be refused access to a door by adding it to the `forbidden` list using
``` 
Doors.Revoke(<door name>)
```

### `Query`

The `Query` entity is placeholder function for future requirements that may require
more information than is provided in the `Cards` objects. At the moment it implements
a single convenience query:

- `HasGroup(groups []string, group string)`, where:
   - `groups` is a list of groups (typically a card's assigned groups) 
   - `group` is the group for which to check

e.g.:
```
QUERY.HasGroup(CARD.Groups,"Student")
```

## Sample `acl.grules` file

```
rule HarryPotter "Harry Potter can sneak into Dungeon" {
     when
         CARD.Name == "Harry Potter"
     then
         DOORS.Allow("Dungeon");
         Retract("HarryPotter");
}

rule Hermione "Hermione is allowed in the Kitchen because House Elves" {
     when
         CARD.Name == "Hermione Granger"
     then
         DOORS.Allow("Kitchen");
         Retract("Hermione");
}


rule Teachers "Teachers have access to all houses" {
     when
         QUERY.HasGroup(CARD.Groups,"Teacher")
     then
         DOORS.Allow("Gryffindor");
         DOORS.Allow("Hufflepuff");
         DOORS.Allow("Ravenclaw");
         DOORS.Allow("Slytherin");
         Retract("Teachers");
}

rule Hogsmeade "Students do not have access to Hogsmeade on weekdays" {
      when
         QUERY.HasGroup(CARD.Groups,"Student") && Now().Format("Monday") != "Saturday" && Now().Format("Monday") != "Sunday"
     then
         DOORS.Forbid("Hogsmeade");
         Retract("Hogsmeade");
}
```