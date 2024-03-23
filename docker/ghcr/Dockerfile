FROM alpine

EXPOSE 8080
EXPOSE 8443

RUN mkdir -p /usr/local/etc/uhppoted/httpd
RUN mkdir -p /usr/local/etc/uhppoted/httpd/grules
RUN mkdir -p /usr/local/etc/uhppoted/httpd/system
RUN mkdir -p /usr/local/etc/uhppoted/httpd/audit

COPY uhppoted.conf /usr/local/etc/uhppoted
COPY auth.json     /usr/local/etc/uhppoted/httpd
COPY acl.grl       /usr/local/etc/uhppoted/httpd

ADD grules  /usr/local/etc/uhppoted/httpd/grules
ADD system  /usr/local/etc/uhppoted/httpd/system
ADD html    /usr/local/etc/uhppoted/httpd/html

WORKDIR /opt/uhppoted 
COPY uhppoted-httpd .

ENTRYPOINT /opt/uhppoted/uhppoted-httpd --debug --config /usr/local/etc/uhppoted/uhppoted.conf --console
