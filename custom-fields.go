package trello

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/pkg/errors"
)

//  //  //

// CustomFieldItem represents the custom field items of Trello a trello card.
type CustomFieldItem struct {
	ID            string           `json:"id,omitempty"`
	Value         CustomFieldValue `json:"value,omitempty"`
	IDValue       string           `json:"idValue,omitempty"`
	IDCustomField string           `json:"idCustomField,omitempty"`
	IDModel       string           `json:"idModel,omitempty"`
	IDModelType   string           `json:"modelType,omitempty"`
}

func (c *Client) SetCustomFieldByItem(cfi CustomFieldItem, extraArgs ...Arguments) error {
	if cfi.IDModelType != "" && cfi.IDModelType != "card" {
		return errors.Errorf("unsupported model type: %s", cfi.IDModelType)
	}

	path := fmt.Sprintf("cards/%s/customField/%s/item", cfi.IDModel, cfi.IDCustomField)
	args := flattenArguments(extraArgs)
	cfValue := CustomFieldItem{
		Value: cfi.Value,
	}

	return c.PutJSON(path, args, cfValue, nil)
}

func (c *Client) SetCustomField(cardID, customFieldID string, value any, extraArgs ...Arguments) error {
	path := fmt.Sprintf("cards/%s/customField/%s/item", cardID, customFieldID)
	args := flattenArguments(extraArgs)
	cfValue := CustomFieldItem{
		Value: NewCustomFieldValue(value),
	}

	return c.PutJSON(path, args, cfValue, nil)
}

// CustomFieldValue represents the custom field value struct
type CustomFieldValue struct {
	val interface{}
}

type cfval struct {
	Text    string `json:"text,omitempty"`
	Number  string `json:"number,omitempty"`
	Date    string `json:"date,omitempty"`
	Checked string `json:"checked,omitempty"`
}

// NewCustomFieldValue the custom field constructor
func NewCustomFieldValue(val interface{}) CustomFieldValue {
	return CustomFieldValue{val: val}
}

const timeFmt = "2006-01-02T15:04:05Z"

// Get the custom field value getter
func (v CustomFieldValue) Get() interface{} {
	return v.val
}

// String the custom field String method
func (v CustomFieldValue) String() string {
	return fmt.Sprintf("%s", v.val)
}

// MarshalJSON the custom field marchaller
func (v CustomFieldValue) MarshalJSON() ([]byte, error) {
	val := v.val

switchVal:
	switch v := val.(type) {
	case driver.Valuer:
		var err error
		val, err = v.Value()
		if err != nil {
			return nil, err
		}
		goto switchVal
	case string:
		if v == "" {
			return json.Marshal("")
		}

		return json.Marshal(cfval{Text: v})
	case int, int64:
		return json.Marshal(cfval{Number: fmt.Sprintf("%d", v)})
	case float64:
		return json.Marshal(cfval{Number: fmt.Sprintf("%f", v)})
	case bool:
		if v {
			return json.Marshal(cfval{Checked: "true"})
		}
		return json.Marshal(cfval{Checked: "false"})
	case time.Time:
		return json.Marshal(cfval{Date: v.Format(timeFmt)})
	case cfval:
		return json.Marshal(v)
	default:
		return nil, fmt.Errorf("unsupported type")
	}
}

// UnmarshalJSON the custom field umarshaller
func (v *CustomFieldValue) UnmarshalJSON(b []byte) error {
	cfval := cfval{}
	err := json.Unmarshal(b, &cfval)
	if err != nil {
		return err
	}
	if cfval.Text != "" {
		v.val = cfval.Text
	}
	if cfval.Date != "" {
		v.val, err = time.Parse(timeFmt, cfval.Date)
		if err != nil {
			return err
		}
	}
	if cfval.Checked != "" {
		v.val = cfval.Checked == "true"
	}
	if cfval.Number != "" {
		v.val, err = strconv.Atoi(cfval.Number)
		if err != nil {
			v.val, err = strconv.ParseFloat(cfval.Number, 64)
			if err != nil {
				v.val, err = strconv.ParseFloat(cfval.Number, 32)
				if err != nil {
					v.val, err = strconv.ParseInt(cfval.Number, 10, 64)
					if err != nil {
						return fmt.Errorf("cannot convert %s to number", cfval.Number)
					}
				}
			}
		}
	}
	return nil
}

// CustomField represents Trello's custom fields: "extra bits of structured data
// attached to cards when our users need a bit more than what Trello provides."
// https://developers.trello.com/reference/#custom-fields
type CustomField struct {
	ID          string `json:"id"`
	IDModel     string `json:"idModel"`
	IDModelType string `json:"modelType,omitempty"`
	FieldGroup  string `json:"fieldGroup"`
	Name        string `json:"name"`
	Pos         int    `json:"pos"`
	Display     struct {
		CardFront bool `json:"cardfront"`
	} `json:"display"`
	Type    string               `json:"type"`
	Options []*CustomFieldOption `json:"options"`
}

// CustomFieldOption are nested resources of CustomFields
type CustomFieldOption struct {
	ID            string `json:"id"`
	IDCustomField string `json:"idCustomField"`
	Value         struct {
		Text string `json:"text"`
	} `json:"value"`
	Color string `json:"color,omitempty"`
	Pos   int    `json:"pos"`
}

// GetCustomField takes a field id string and Arguments and returns the matching custom Field.
func (c *Client) GetCustomField(fieldID string, extraArgs ...Arguments) (customField *CustomField, err error) {
	args := flattenArguments(extraArgs)
	path := fmt.Sprintf("customFields/%s", fieldID)
	err = c.Get(path, args, &customField)
	return
}

// GetCustomFields returns a slice of all receiver board's custom fields.
func (b *Board) GetCustomFields(extraArgs ...Arguments) (customFields []*CustomField, err error) {
	args := flattenArguments(extraArgs)
	path := fmt.Sprintf("boards/%s/customFields", b.ID)
	err = b.client.Get(path, args, &customFields)
	return
}
