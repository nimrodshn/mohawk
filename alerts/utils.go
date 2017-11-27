package alerts

func (alert *Alert) checkOutOfRange(value float64) bool {
	var res bool
	switch alert.Range.Type {
	case BETWEEN:
		if (value < alert.Range.Min || value > alert.Range.Max) {
			res = true
		} else {
			res = false
		}
	case LOWER_THAN:
		if (value > alert.Range.Max) {
			res = true
		} else {
			res =  false
		}
	case HIGHER_THAN:
		if (value < alert.Range.Min) {
			res = true
		} else {
			res = false
		}
	}
	return res
}

func (alert *Alert) setStatus(status bool){
	alert.State = status
}

func (Alerts Alerts) Open() {}


