package catalog

type Objects struct {
	objects []Object
}

func (o *Objects) Append(list ...Object) {
	if list != nil {
		o.objects = append(o.objects, list...)
	}
}

func (o Objects) AsArray() []Object {
	return o.objects
}
