package entity

type BasicRecord struct {
	ID   int               `json:"id"`
	Data map[string]string `json:"data"`
}

func (d *BasicRecord) Copy() BasicRecord {
	values := d.Data

	newMap := map[string]string{}
	for key, value := range values {
		newMap[key] = value
	}

	return BasicRecord{
		ID:   d.ID,
		Data: newMap,
	}
}
