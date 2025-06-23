package H

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var time_layouts = []string{
	time.RFC3339,
	time.RFC3339Nano,
	"2006-01-02 15:04:05.000000",
	"2006-01-02T15:04:05.000000",
	"2006-01-02 15:04:05",
	"2006-01-02T15:04:05",
	"2006-01-02 15:04",
	"2006-01-02T15:04",
	"2006/01/02 15:04:05",
	"2006/01/02 15:04",
	time.RFC822,
	time.RFC850,
}

func StructToMap(s interface{}, clean_string bool) map[string]interface{} {
	v := reflect.ValueOf(s)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil
	}
	m := make(map[string]interface{})
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("json")
		if tag == "-" || IsEmpty(tag) {
			continue
		}
		parts := strings.Split(tag, ",")
		name := parts[0]
		if IsEmpty(name) {
			continue
		}

		value := v.Field(i).Interface()
		if clean_string {
			switch v := value.(type) {
			case string:
				if !IsEmpty(v) {
					value = RemoveNonPrintable(v)
				}
			case *string:
				if !IsEmpty(v) {
					cleaned := RemoveNonPrintable(*v)
					value = &cleaned
				}
			}
		}
		m[name] = value
	}
	return m
}

func MapToStruct(data map[string]interface{}, result *interface{}) bool {
	v := reflect.ValueOf(*result)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return false
	}
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("json")
		if tag == "-" || IsEmpty(tag) {
			continue
		}
		parts := strings.Split(tag, ",")
		name := parts[0]
		if IsEmpty(name) {
			continue
		}
		if current, ok := data[tag]; ok {
			v.Field(i).Set(reflect.ValueOf(current))
		}
	}
	return true
}

func StringToTime(date string) *time.Time {
	if !IsEmpty(date) {
		if t, err := time.Parse(time.RFC3339, date); err == nil {
			return &t
		}
	}
	return nil
}
func StringToDate(date string) *time.Time {
	if !IsEmpty(date) {
		if !strings.Contains(date, "T") {
			date = date + "T12:00:00Z"
		}
		if t, err := time.Parse(time.RFC3339, date); err == nil {
			return &t
		}
	}
	return nil
}
func StringToDateEndTime(date string) *time.Time {
	if !IsEmpty(date) {
		if !strings.Contains(date, "T") {
			date = date + "T23:59:59Z"
		}
		if t, err := time.Parse(time.RFC3339, date); err == nil {
			return &t
		}
	}
	return nil
}

func StringDateTimeToDate(date string) *time.Time {
	if !IsEmpty(date) {
		if t, err := time.Parse(time.RFC3339, date); err == nil {
			return &t
		}
	}
	return nil
}

func Float64ToString(inputNum float64, decimals uint) string {
	return strconv.FormatFloat(Round(inputNum, decimals), 'f', -1, 64)
}

func IntToString(inputNum int) string {
	return strconv.Itoa(inputNum)
}

func UIntToString(inputNum uint) string {
	return strconv.Itoa(int(inputNum))
}

func Int64ToString(inputNum int64) string {
	return strconv.FormatInt(inputNum, 10)
}

func Uint64ToString(inputNum uint64) string {
	return strconv.FormatUint(inputNum, 10)
}

// Time to YYYY-MM-DD HH:MM
func TimeToString(t time.Time) string {
	return t.Format("2006-01-02 15:04")
}

func TimeToDateString(t time.Time) string {
	return t.Format("2006-01-02")
}

func TimeToDateTimeString(t time.Time) string {
	// YYYY-MM-DD HH:MM
	return t.Format("2006-01-02 15:04")
}

func TimeToSeniatTimeString(t time.Time, is_utc bool) string {
	// DD-MM-YYYY H:MM:SS AM/PM
	if is_utc {
		return t.Add(-4 * time.Hour).Format("02-01-2006 3:04:05 PM")
	}
	return t.Format("02-01-2006 3:04:05 PM")
}

func TimeToTimeString(t time.Time) string {
	// HH:MM
	return t.Format("15:04")
}

func TimeToFactoryDigitalDate(t time.Time) string {
	// dd/MM/AAAA
	return t.Format("02/01/2006")
}

func TimeToFullTimeString(t time.Time) string {
	// hh:mm:ss tt
	return strings.ToLower(t.Format("03:04:05 PM"))
}

// TBool is a custom boolean type that can be compared with bool and assigned from/to bool
type TBool bool

func (t TBool) ToBool() bool {
	return bool(t)
}

func (t TBool) Equal(b bool) bool {
	return bool(t) == b
}

func (s TBool) Value() (driver.Value, error) {
	if bool(s) {
		return uint64(1), nil
	}
	return uint64(0), nil
}

func (s *TBool) Scan(value interface{}) error {
	if value == nil {
		*s = false
		return nil
	}

	switch v := value.(type) {
	case bool:
		*s = TBool(v)
	case int64:
		*s = v == 1
	case []byte:
		*s = string(v) == "1" || string(v) == "true"
	case string:
		*s = v == "1" || v == "true"
	default:
		return fmt.Errorf("cannot convert %T to bool", value)
	}
	return nil
}

func (s *TBool) UnmarshalJSON(data []byte) error {
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	switch x := v.(type) {
	case bool:
		*s = TBool(x)
	case float64:
		*s = TBool(x > 0.00001 || x < -0.00001)
	case int64:
		*s = TBool(x == 1)
	case uint:
		*s = TBool(x == 1)
	case uint64:
		*s = TBool(x == 1)
	case int:
		*s = TBool(x == 1)
	case string:
		*s = TBool(!(x == "" || x == "0" || x == "false"))
	default:
		*s = TBool(false)
	}
	return nil
}

type TTime time.Time

func (t TTime) Value() (driver.Value, error) {
	return time.Time(t), nil
}
func (ct *TTime) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	if s == "null" || s == "" {
		*ct = TTime(time.Time{})
		return nil
	}

	var err error
	for _, layout := range time_layouts {
		t, err := time.Parse(layout, s)
		if err == nil {
			*ct = TTime(t)
			return nil
		}
	}
	return err
}

func (ct TTime) MarshalJSON() ([]byte, error) {
	if time.Time(ct).IsZero() {
		return []byte("null"), nil
	}

	return []byte(fmt.Sprintf("\"%s\"", time.Time(ct).Format(time.RFC3339))), nil
}

func (ct *TTime) Scan(value interface{}) error {
	if value == nil {
		*ct = TTime(time.Time{})
		return nil
	}

	switch v := value.(type) {
	case time.Time:
		*ct = TTime(v)

	case string:
		for _, layout := range time_layouts {
			t, err := time.Parse(layout, v)
			if err == nil {
				*ct = TTime(t)
				return nil
			}
		}

	case []byte:
		for _, layout := range time_layouts {
			t, err := time.Parse(layout, string(v))
			if err == nil {
				*ct = TTime(t)
				return nil
			}
		}
	}
	return nil
}

func (t *TTime) ToTime() time.Time {
	if t == nil {
		return time.Time{}
	}
	return time.Time(*t)
}

type TDeletedAt struct {
	gorm.DeletedAt
}

func (t TDeletedAt) Value() (driver.Value, error) {
	if t.DeletedAt.Time.IsZero() {
		return nil, nil
	}
	return t.DeletedAt.Time, nil
}

func (t *TDeletedAt) Scan(value interface{}) error {
	if value == nil {
		t.DeletedAt = gorm.DeletedAt{}
		return nil
	}

	switch v := value.(type) {
	case time.Time:
		t.DeletedAt = gorm.DeletedAt{Time: v, Valid: true}
	case []byte:
		for _, layout := range time_layouts {
			parsedTime, err := time.Parse(layout, string(v))
			if err == nil {
				t.DeletedAt = gorm.DeletedAt{Time: parsedTime, Valid: true}
				return nil
			}
		}
	}
	return nil
}

func (t *TDeletedAt) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(data), `"`)
	if s == "null" || s == "" {
		t.DeletedAt = gorm.DeletedAt{}
		return nil
	}

	var err error
	for _, layout := range time_layouts {
		parsedTime, err := time.Parse(layout, s)
		if err == nil {
			t.DeletedAt = gorm.DeletedAt{Time: parsedTime, Valid: true}
			return nil
		}
	}
	return err
}

func (t TDeletedAt) MarshalJSON() ([]byte, error) {
	if t.DeletedAt.Time.IsZero() {
		return []byte("null"), nil
	}
	return []byte(fmt.Sprintf("\"%s\"", t.DeletedAt.Time.Format(time.RFC3339))), nil
}

func (t *TDeletedAt) ToTime() time.Time {
	if t == nil || t.DeletedAt.Time.IsZero() {
		return time.Time{}
	}
	return t.DeletedAt.Time
}

func (t TDeletedAt) GormValue(ctx context.Context, db *gorm.DB) clause.Expr {
	if !t.Valid {
		return gorm.Expr("NULL")
	}
	return gorm.Expr("?", t.Time)
}
