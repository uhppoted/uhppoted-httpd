FROM alpine

EXPOSE 8888

RUN mkdir -p /usr/local/etc/uhppoted/httpd
RUN mkdir -p /usr/local/etc/uhppoted/httpd/grules
RUN mkdir -p /usr/local/etc/uhppoted/httpd/system
RUN mkdir -p /usr/local/etc/uhppoted/httpd/audit

COPY uhppoted.conf           /usr/local/etc/uhppoted
COPY auth.json               /usr/local/etc/uhppoted/httpd

COPY acl.grl                 /usr/local/etc/uhppoted/httpd
COPY grules/interfaces.grl   /usr/local/etc/uhppoted/httpd/grules
COPY grules/controllers.grl  /usr/local/etc/uhppoted/httpd/grules
COPY grules/doors.grl        /usr/local/etc/uhppoted/httpd/grules
COPY grules/groups.grl       /usr/local/etc/uhppoted/httpd/grules
COPY grules/cards.grl        /usr/local/etc/uhppoted/httpd/grules
COPY grules/users.grl        /usr/local/etc/uhppoted/httpd/grules
COPY grules/events.grl       /usr/local/etc/uhppoted/httpd/grules
COPY grules/logs.grl         /usr/local/etc/uhppoted/httpd/grules

COPY system/interfaces.json  /usr/local/etc/uhppoted/httpd/system
COPY system/controllers.json /usr/local/etc/uhppoted/httpd/system
COPY system/doors.json       /usr/local/etc/uhppoted/httpd/system
COPY system/groups.json      /usr/local/etc/uhppoted/httpd/system
COPY system/cards.json       /usr/local/etc/uhppoted/httpd/system
COPY system/users.json       /usr/local/etc/uhppoted/httpd/system
COPY system/events.json      /usr/local/etc/uhppoted/httpd/system
COPY system/history.json     /usr/local/etc/uhppoted/httpd/system
COPY system/logs.json        /usr/local/etc/uhppoted/httpd/system


WORKDIR /opt/uhppoted 
COPY uhppoted-httpd .

ENTRYPOINT /opt/uhppoted/uhppoted-httpd --debug --config /usr/local/etc/uhppoted/uhppoted.conf --console
