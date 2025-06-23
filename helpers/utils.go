package H

import (
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/labstack/echo/v4"
	"github.com/leekchan/accounting"
	"gorm.io/gorm"

	"github.com/google/uuid"
)

func InArray[T comparable](val T, array []T) (exists bool, index int) {
	for i, v := range array {
		if v == val {
			return true, i
		}
	}
	return false, -1
}

func Round(num float64, decimals uint) float64 {
	var rounder float64
	pow := math.Pow(10, float64(decimals))
	intermed := num * pow
	_, div := math.Modf(intermed)
	if div >= 0.5 {
		rounder = math.Ceil(intermed)
	} else {
		rounder = math.Floor(intermed)
	}
	return rounder / pow
}

var last_id_generator uint = 0
var mutex_id_generator = &sync.Mutex{}

func JSONEncode(data interface{}) string {
	bytes, _ := json.Marshal(data)
	return string(bytes)
}
func JSONEncodePtr(data interface{}) *string {
	result := JSONEncode(data)
	return &result
}
func JSONEncodePretty(data interface{}) string {
	bytes, _ := json.MarshalIndent(data, "", "    ")
	return string(bytes)
}

func JSONDecode(data string) interface{} {
	var result interface{}
	err := json.Unmarshal([]byte(data), &result)
	if err != nil {
		return nil
	}
	return result
}
func JSONDecodeMap(data string) map[string]interface{} {
	var result map[string]interface{}
	err := json.Unmarshal([]byte(data), &result)
	if err != nil {
		return nil
	}
	return result
}
func JSONDecodeMapPtr(data string) *map[string]interface{} {
	r := JSONDecodeMap(data)
	if r == nil {
		return nil
	}
	return &r
}

func NewUUID() string {
	// Generate a random UUID
	v := uuid.New().String()

	// Remove the last 17 characters
	v2 := v[0 : len(v)-17]

	// Generate a hex string from the current time
	now := time.Now().UTC().Unix()
	now_hex := fmt.Sprintf("%x", now)

	// Generate a hex string from a global counter
	mutex_id_generator.Lock()
	last_id_generator += 1
	if last_id_generator > 65535 {
		last_id_generator = 0
	}
	last_id_generator_hex := fmt.Sprintf("%04x", last_id_generator)
	mutex_id_generator.Unlock()

	// Concatenate the first part of the UUID with the time and counter
	v3 := v2 + last_id_generator_hex + "-"

	// Return the first part of the UUID with the time and counter\
	first := len(v3) + len(now_hex)
	diff := len(v) - first
	return v3 + v[first:first+diff] + now_hex
}

func UrlDecode(query string) string {
	if d, err := url.QueryUnescape(query); err == nil {
		return d
	}
	return query
}

// Translation is a structure to store translations in the JSON file
type Translation map[string]string
type TranslationCache map[string]Translation

var translationCache TranslationCache

func GetLanguage(c echo.Context) string {
	language := c.Request().Header.Get("X-Language")
	if IsEmpty(language) {
		language = "es" // Default value if no language is specified
	}
	switch language {
	case "es", "en", "ES", "EN":
		return strings.ToLower(language)
	default:
		return "es"
	}
}

func TranslateText(text string, c echo.Context) string {

	language := GetLanguage(c)
	if language == "en" {
		return text
	}
	// Get the root directory of the project
	projectRoot, err := os.Getwd()
	if err != nil {
		return text
	}
	jsonFile := filepath.Join(projectRoot, "translations/"+language+".json")
	if translationCache == nil {
		translationCache = make(map[string]Translation)
	}
	inCache, ok := translationCache[language]
	if !ok {
		var translations Translation
		// Build the path to the language JSON file

		// Check if the JSON file exists
		_, err = os.Stat(jsonFile)
		if os.IsNotExist(err) {
			// The JSON file does not exist, return the original text without translation
			return text
		} else if err != nil {
			// An error occurred while checking the existence of the file
			return text
		}

		// Read the language JSON file
		jsonData, err := os.ReadFile(jsonFile)
		if err != nil {
			return text
		}

		err = json.Unmarshal(jsonData, &translations)
		if err != nil {
			return text
		}

		translationCache[language] = translations

		inCache = translationCache[language]
	}
	// Search for the text in the translations
	translatedText, ok := inCache[text]
	if !ok {
		// The text was not found in the translations, return the original text without translation
		textTranslated := TranslateTextWithIA(text, "en", language)
		translationCache[language][text] = textTranslated
		jsonToWrite := JSONEncodePretty(translationCache[language])
		_ = os.WriteFile(jsonFile, []byte(jsonToWrite), 0644)

		return textTranslated
	}

	return translatedText
}

func CleanRif(rif string) string {
	rif = strings.ToUpper(rif)

	re := regexp.MustCompile("[^VEJFG0-9]+")
	return re.ReplaceAllString(rif, "")
}

// IsEmpty checks if a value is considered empty based on its type
//
// Parameters:
//   - s: interface{} - The value to check, can be of various types:
//   - string or *string - Checks if nil or empty string after trimming
//   - time.Time or *time.Time - Checks if zero time
//   - gorm.DeletedAt or *gorm.DeletedAt - Checks if not valid
//   - Arrays/Slices/Maps - Checks if length is 0
//   - Other pointer types - Checks if nil
//
// Returns:
//   - bool - true if the value is considered empty, false otherwise
//
// Examples:
//
//	IsEmpty("")      // Returns true
//	IsEmpty("test")  // Returns false
//	IsEmpty(nil)     // Returns true
//	IsEmpty([]{})    // Returns true
//	IsEmpty([1,2])   // Returns false
func IsEmpty(s interface{}) bool {
	switch v := s.(type) {
	case string:
		if Trim(v) == "" {
			return true
		}
	case time.Time:
		if v.IsZero() {
			return true
		}
	case gorm.DeletedAt:
		if !v.Valid || v.Time.IsZero() {
			return true
		}
	case TTime:
		if v.ToTime().IsZero() {
			return true
		}
	case TDeletedAt:
		if !v.Valid || v.ToTime().IsZero() {
			return true
		}
	case TBool:
		if !v.ToBool() {
			return true
		}
	case bool:
		if !v {
			return true
		}
	default:
		reflectValue := reflect.ValueOf(s)
		if reflectValue.Kind() == reflect.Ptr {
			if reflectValue.IsNil() {
				return true
			}
			reflectValue = reflectValue.Elem()
			return IsEmpty(reflectValue.Interface())
		}
		switch reflectValue.Kind() {
		case reflect.Array, reflect.Slice, reflect.Map:
			return reflectValue.Len() == 0
		}
	}
	return false
}

func GetIP(c echo.Context) string {
	client_ip_string := c.Request().Header.Get("CF-Connecting-IP")

	if client_ip_string == "" {
		client_ip_string = c.RealIP()
	}

	if client_ip_string == "" || client_ip_string == "::1" || client_ip_string == "127.0.0.1" {
		client_ip_string = "38.196.222.29"
	}
	return client_ip_string
}

func IsFloatEmpty(f *float64) bool {
	return f == nil || math.Abs(*f) < 0.000001
}

func IsTimeEmpty(t *time.Time) bool {
	if t == nil {
		return true
	}
	return t.IsZero()
}

func VenezuelaGetValidRif(typeStr string, ci string) string {
	ci = strings.ToUpper(RemoveNonNumeric(ci))
	if len(ci) > 9 {
		return ""
	}

	countDigits := len(ci)
	if countDigits == 9 {
		countDigits--
	}

	calc := [10]int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	constants := [9]int{4, 3, 2, 7, 6, 5, 4, 3, 2}

	switch typeStr {
	case "V":
		calc[0] = 1
	case "E":
		calc[0] = 2
	case "J":
		calc[0] = 3
	case "P":
		calc[0] = 4
	case "G":
		calc[0] = 5
	default:
		return ""
	}

	sum := calc[0] * constants[0]
	index := len(constants) - 1

	for i := countDigits - 1; i >= 0; i-- {
		digit, err := strconv.Atoi(string(ci[i]))
		if err != nil {
			return ""
		}
		calc[index] = digit
		sum += digit * constants[index]
		index--
	}

	finalDigit := sum % 11
	if finalDigit > 1 {
		finalDigit = 11 - finalDigit
	}

	if len(ci) == 9 {
		finalDigitLegal, err := strconv.Atoi(string(ci[8]))
		if err != nil {
			return ""
		}
		if finalDigitLegal != finalDigit && finalDigitLegal != 0 {
			return ""
		}
		calc[9] = finalDigitLegal
	} else {
		calc[9] = finalDigit
	}

	rif := typeStr
	for i := 1; i < len(calc); i++ {
		rif += strconv.Itoa(calc[i])
	}

	return rif
}

func ValidRecaptchaV2(recaptchaToken string) bool {
	if IsEmpty(recaptchaToken) {
		return false
	}
	url := "https://www.google.com/recaptcha/api/siteverify"
	request := map[string]string{
		"secret":   os.Getenv("RECAPTCHA_SECRET"),
		"response": recaptchaToken,
	}

	postFormBody := ""
	for k, v := range request {
		postFormBody += k + "=" + v + "&"
	}
	resp, err := http.Post(url, "application/x-www-form-urlencoded", strings.NewReader(postFormBody))
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false
	}

	result := JSONDecode(string(body)).(map[string]interface{})

	if result["success"].(bool) {
		return result["hostname"].(string) == os.Getenv("RECAPTCHA_VALID_HOSTNAME")
	}
	return false
}

func ArgentinaValidateCUIT(cuit string) bool {
	cuit = RemoveNonNumeric(cuit)
	if len(cuit) != 11 {
		return false
	}
	baseWeights := []int{5, 4, 3, 2, 7, 6, 5, 4, 3, 2}

	checksum := 0
	for i := 0; i < 10; i++ {
		digit := int(cuit[i] - '0')
		checksum += baseWeights[i] * digit
	}

	checkDigit := 11 - (checksum % 11)
	validDigit := checkDigit
	if checkDigit == 11 {
		validDigit = 0
	} else if checkDigit == 10 {
		validDigit = 9
	}

	lastDigit := int(cuit[10] - '0')

	return lastDigit == validDigit
}

func ArgentinaValidateCUIL(cuit string) bool {
	cuit = RemoveNonNumeric(cuit)
	if len(cuit) != 11 {
		return false
	}
	digits := make([]int, 11)
	for i := 0; i < 11; i++ {
		digits[i] = int(cuit[i] - '0')
	}

	checkDigit := digits[10]
	rest := digits[:10]

	total := 0
	for i, digit := range rest {
		total += digit * (2 + (i % 6))
	}

	mod11 := 11 - (total % 11)

	if mod11 == 11 {
		return checkDigit == 0
	}

	if mod11 == 10 {
		return false
	}

	return checkDigit == mod11
}

type ReportRecord struct {
	Data            string   `json:"data" example:"1.000,00"`
	NumericData     *float64 `json:"numeric_data" example:"1000.00"`
	ColumnKey       string   `json:"column_key" example:"total"`
	ColSpan         uint     `json:"col_span" example:"0"`
	Bold            bool     `json:"bold"`
	Italic          bool     `json:"italic"`
	Underline       bool     `json:"underline"`
	AlignmentCenter bool     `json:"alignment_center"`
	AlignmentRight  bool     `json:"alignment_right"`
	ColorRGB        string   `json:"color_rgb" example:"#ff0000"`
	IsNumeric       bool     `json:"is_numeric"`
}

type ReportFormat struct {
	Title   string           `json:"title"`
	Columns []ReportRecord   `json:"columns"`
	Rows    [][]ReportRecord `json:"rows"`
}

func DivideSkuList(skuList []string) (string, []string) {
	sku := ""
	var other_skus []string
	if len(skuList) > 0 {
		sku = skuList[0]
		if len(skuList) > 1 {
			for j := 1; j < len(skuList); j++ {
				other_skus = append(other_skus, skuList[j])
			}
		}
	}

	return sku, other_skus
}

func MaybeFormatNumber(number float64, formatted bool) string {
	if formatted {
		result := strings.TrimRight(accounting.FormatNumber(number, 4, ".", ","), "0")
		if IsEmpty(result) {
			return "0"
		}

		if result[len(result)-1] == ',' {
			result = strings.TrimRight(result, ",")
			if result == "-0" {
				return "0"
			}
		}

		if len(result) > 2 && result[len(result)-2] == ',' {
			return result + "0"
		}
		if result == "-0" {
			return "0"
		}
		return result
	}
	result := strings.TrimRight(accounting.FormatNumber(number, 4, "", "."), "0")
	if result[len(result)-1] == '.' {
		result = strings.TrimRight(result, ".")
	}
	if result == "-0" {
		return "0"
	}
	return result
}

func Contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// RemoveNonNumeric removes all non-numeric characters from a string.
//
// This function takes a string and removes all characters that are not
// digits from it.
//
// Parameters:
// - s: The string from which non-numeric characters should be removed.
//
// Returns:
// - The modified string with all non-numeric characters removed.
func RemoveNonNumeric(s string) string {
	// Use a regular expression to match all characters that are not digits.
	re := regexp.MustCompile("[^0-9]+")
	// Replace all matches with an empty string.
	return re.ReplaceAllString(s, "")
}

func IsLetter(s string) bool {
	return regexp.MustCompile("^[a-zA-Z]+$").MatchString(s)
}

func RemoveNonNumericAndLetters(s string) string {
	re := regexp.MustCompile("[^a-zA-Z0-9]+")
	return re.ReplaceAllString(s, "")
}
func RemoveNonPrintable(s string) string {
	result := make([]rune, 0, len(s))
	for _, r := range s {
		if unicode.IsPrint(r) {
			result = append(result, r)
		}
	}
	return string(result)
}

func StringToPtr(s string) *string {
	return &s
}

func StringFromPtr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func StringToUInt64(s string) uint64 {
	result, _ := strconv.ParseUint(s, 10, 64)
	return result
}

func StringToInt64(s string) int64 {
	result, _ := strconv.ParseInt(s, 10, 64)
	return result
}

func StringToFloat64(s string) float64 {
	result, _ := strconv.ParseFloat(s, 64)
	return result
}

func Trim(s string) string {
	return strings.Trim(s, " \t\n\r")
}

func TrimLeft(s string) string {
	return strings.TrimLeft(s, " \t\n\r")
}

func TrimRight(s string) string {
	return strings.TrimRight(s, " \t\n\r")
}

func Validate(data interface{}, c echo.Context) error {
	v := reflect.ValueOf(data)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return errors.New(TranslateText("Invalid data structure", c))
	}
	ConvertEmptyStringsToNil(data)
	return c.Validate(ValidateStruct{Context: c, Data: data})
}

func ConvertEmptyStringsToNil(data interface{}) {
	v := reflect.ValueOf(data)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return
	}
	v = v.Elem()
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if field.Kind() == reflect.Ptr && field.Type().Elem().Kind() == reflect.String {
			if field.IsNil() {
				continue
			}
			strValue := field.Elem().String()
			if IsEmpty(strValue) {
				field.Set(reflect.Zero(field.Type()))
			}
		} else if field.Kind() == reflect.Slice {
			for j := 0; j < field.Len(); j++ {
				if field.Index(j).Kind() == reflect.Ptr && field.Index(j).Type().Elem().Kind() == reflect.String {
					if field.Index(j).IsNil() {
						continue
					}
					strValue := field.Index(j).Elem().String()
					if IsEmpty(strValue) {
						field.Index(j).Set(reflect.Zero(field.Index(j).Type()))
					}
				} else if field.Index(j).Kind() == reflect.Struct {
					ConvertEmptyStringsToNil(field.Index(j).Addr().Interface())
				} else if field.Index(j).Kind() == reflect.Ptr && field.Index(j).Type().Elem().Kind() == reflect.Struct {
					ConvertEmptyStringsToNil(field.Index(j).Interface())
				}
			}
		} else if field.Kind() == reflect.Map {
			for _, value := range field.MapKeys() {
				if field.MapIndex(value).Kind() == reflect.Ptr && field.MapIndex(value).Type().Elem().Kind() == reflect.String {
					if field.MapIndex(value).IsNil() {
						continue
					}
					strValue := field.MapIndex(value).Elem().String()
					if IsEmpty(strValue) {
						field.MapIndex(value).Set(reflect.Zero(field.MapIndex(value).Type()))
					}
				} else if field.MapIndex(value).Kind() == reflect.Struct {
					ConvertEmptyStringsToNil(field.MapIndex(value).Addr().Interface())
				} else if field.MapIndex(value).Kind() == reflect.Ptr && field.MapIndex(value).Type().Elem().Kind() == reflect.Struct {
					ConvertEmptyStringsToNil(field.MapIndex(value).Interface())
				}
			}
		} else if field.Kind() == reflect.Ptr && field.Type().Elem().Kind() == reflect.Struct {
			ConvertEmptyStringsToNil(field.Interface())
		} else if field.Kind() == reflect.Struct {
			ConvertEmptyStringsToNil(field.Addr().Interface())
		}
	}
}

// SliceGetRecursive recursively traverses a slice using an array of keys to access nested values
// It can handle nested maps and slices
// Example:
// slice := []interface{}{map[string]interface{}{"foo": []interface{}{1, 2, 3}}}
// SliceGetRecursive(slice, []string{"0", "foo", "1"}) // Returns 2
func SliceGetRecursive(m []interface{}, key []string) interface{} {
	if len(key) == 0 {
		return m
	}
	index, _ := strconv.Atoi(key[0])
	if index >= 0 && index < len(m) {
		if len(key) == 1 {
			return m[index]
		} else if reflect.TypeOf(m[index]).Kind() == reflect.Map {
			return MapGetRecursive(m[index].(map[string]interface{}), key[1:])
		} else if reflect.TypeOf(m[index]).Kind() == reflect.Slice {
			return SliceGetRecursive(m[index].([]interface{}), key[1:])
		}
	}
	return nil
}

// MapGetRecursive recursively traverses a map using an array of keys to access nested values
// It can handle nested maps and slices
// Example:
// m := map[string]interface{}{"foo": map[string]interface{}{"bar": "baz"}}
// MapGetRecursive(m, []string{"foo", "bar"}) // Returns "baz"
func MapGetRecursive(m map[string]interface{}, key []string) interface{} {
	if len(key) == 0 {
		return m
	}
	if value, ok := m[key[0]]; ok {
		if len(key) == 1 {
			return value
		} else if reflect.TypeOf(value).Kind() == reflect.Map {
			return MapGetRecursive(value.(map[string]interface{}), key[1:])
		} else if reflect.TypeOf(value).Kind() == reflect.Slice {
			return SliceGetRecursive(value.([]interface{}), key[1:])
		}
	}
	return nil
}

// MapGet gets a value from a map using a dot-notation string key
// The key can access nested maps and slices
// Example:
// m := map[string]interface{}{"foo": map[string]interface{}{"bar": "baz"}}
// MapGet(m, "foo.bar") // Returns "baz"
func MapGet(m map[string]interface{}, key string) interface{} {
	parts := strings.Split(key, ".")
	return MapGetRecursive(m, parts)
}

// SliceGet gets a value from a slice using a dot-notation string key
// The key can access nested maps and slices
// Example:
// slice := []interface{}{map[string]interface{}{"foo": "bar"}}
// SliceGet(slice, "0.foo") // Returns "bar"
func SliceGet(m []interface{}, key string) interface{} {
	parts := strings.Split(key, ".")
	return SliceGetRecursive(m, parts)
}

func Sha256(s string) string {
	h := sha256.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

func Sha512(s string) string {
	h := sha512.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

func FloatToFloatPtr(f float64) *float64 {
	return &f
}

func Float64FromPtr(f *float64) float64 {
	if f == nil {
		return 0
	}
	return *f
}

var unidades = []string{"", "uno", "dos", "tres", "cuatro", "cinco", "seis", "siete", "ocho", "nueve"}
var decenas = []string{"", "diez", "veinte", "treinta", "cuarenta", "cincuenta", "sesenta", "setenta", "ochenta", "noventa"}
var centenas = []string{"", "ciento", "doscientos", "trescientos", "cuatrocientos", "quinientos", "seiscientos", "setecientos", "ochocientos", "novecientos"}
var especiales = []string{"diez", "once", "doce", "trece", "catorce", "quince", "dieciséis", "diecisiete", "dieciocho", "diecinueve"}

func ConvertIntergerToWordsSpanish(n int64) string {
	if n == 0 {
		return "cero"
	}

	var partes []string

	if n >= 1_000_000_000 {
		milMillones := n / 1_000_000_000
		partes = append(partes, ConvertIntergerToWordsSpanish(milMillones)+" mil millones")
		n %= 1_000_000_000
	}

	if n >= 1_000_000 {
		millones := n / 1_000_000
		if millones == 1 {
			partes = append(partes, "un millón")
		} else {
			partes = append(partes, ConvertIntergerToWordsSpanish(millones)+" millones")
		}
		n %= 1_000_000
	}

	if n >= 1_000 {
		miles := n / 1_000
		if miles == 1 {
			partes = append(partes, "mil")
		} else {
			partes = append(partes, ConvertIntergerToWordsSpanish(miles)+" mil")
		}
		n %= 1_000
	}

	if n >= 100 {
		cien := n / 100
		if cien == 1 && n%100 == 0 {
			partes = append(partes, "cien")
		} else {
			partes = append(partes, centenas[cien])
		}
		n %= 100
	}

	if n >= 20 {
		decena := n / 10
		unidad := n % 10
		if unidad == 0 {
			partes = append(partes, decenas[decena])
		} else {
			partes = append(partes, decenas[decena]+" y "+unidades[unidad])
		}
	} else if n >= 10 {
		partes = append(partes, especiales[n-10])
	} else if n > 0 {
		partes = append(partes, unidades[n])
	}

	return strings.Join(partes, " ")
}

func ConvertToTextSpanish(numero float64) string {
	numero = Round(numero, 2)
	entero := int64(numero)
	decimal := int64(Round((numero-float64(entero))*100, 0))

	parteEnteraTexto := ConvertIntergerToWordsSpanish(entero)
	parteDecimalTexto := fmt.Sprintf("%02d/100", decimal)

	return fmt.Sprintf("%s con %s céntimos", parteEnteraTexto, parteDecimalTexto)
}

var countryString = `{
	"AF": "Afghanistan",
	"AX": "Aland Islands",
	"AL": "Albania",
	"DZ": "Algeria",
	"AS": "American Samoa",
	"AD": "Andorra",
	"AO": "Angola",
	"AI": "Anguilla",
	"AQ": "Antarctica",
	"AG": "Antigua And Barbuda",
	"AR": "Argentina",
	"AM": "Armenia",
	"AW": "Aruba",
	"AU": "Australia",
	"AT": "Austria",
	"AZ": "Azerbaijan",
	"BS": "Bahamas",
	"BH": "Bahrain",
	"BD": "Bangladesh",
	"BB": "Barbados",
	"BY": "Belarus",
	"BE": "Belgium",
	"BZ": "Belize",
	"BJ": "Benin",
	"BM": "Bermuda",
	"BT": "Bhutan",
	"BO": "Bolivia",
	"BA": "Bosnia And Herzegovina",
	"BW": "Botswana",
	"BV": "Bouvet Island",
	"BR": "Brazil",
	"IO": "British Indian Ocean Territory",
	"BN": "Brunei Darussalam",
	"BG": "Bulgaria",
	"BF": "Burkina Faso",
	"BI": "Burundi",
	"KH": "Cambodia",
	"CM": "Cameroon",
	"CA": "Canada",
	"CV": "Cape Verde",
	"KY": "Cayman Islands",
	"CF": "Central African Republic",
	"TD": "Chad",
	"CL": "Chile",
	"CN": "China",
	"CX": "Christmas Island",
	"CC": "Cocos (Keeling) Islands",
	"CO": "Colombia",
	"KM": "Comoros",
	"CG": "Congo",
	"CD": "Congo, Democratic Republic",
	"CK": "Cook Islands",
	"CR": "Costa Rica",
	"CI": "Cote D\"Ivoire",
	"HR": "Croatia",
	"CU": "Cuba",
	"CY": "Cyprus",
	"CZ": "Czech Republic",
	"DK": "Denmark",
	"DJ": "Djibouti",
	"DM": "Dominica",
	"DO": "Dominican Republic",
	"EC": "Ecuador",
	"EG": "Egypt",
	"SV": "El Salvador",
	"GQ": "Equatorial Guinea",
	"ER": "Eritrea",
	"EE": "Estonia",
	"ET": "Ethiopia",
	"FK": "Falkland Islands (Malvinas)",
	"FO": "Faroe Islands",
	"FJ": "Fiji",
	"FI": "Finland",
	"FR": "France",
	"GF": "French Guiana",
	"PF": "French Polynesia",
	"TF": "French Southern Territories",
	"GA": "Gabon",
	"GM": "Gambia",
	"GE": "Georgia",
	"DE": "Germany",
	"GH": "Ghana",
	"GI": "Gibraltar",
	"GR": "Greece",
	"GL": "Greenland",
	"GD": "Grenada",
	"GP": "Guadeloupe",
	"GU": "Guam",
	"GT": "Guatemala",
	"GG": "Guernsey",
	"GN": "Guinea",
	"GW": "Guinea-Bissau",
	"GY": "Guyana",
	"HT": "Haiti",
	"HM": "Heard Island & Mcdonald Islands",
	"VA": "Holy See (Vatican City State)",
	"HN": "Honduras",
	"HK": "Hong Kong",
	"HU": "Hungary",
	"IS": "Iceland",
	"IN": "India",
	"ID": "Indonesia",
	"IR": "Iran, Islamic Republic Of",
	"IQ": "Iraq",
	"IE": "Ireland",
	"IM": "Isle Of Man",
	"IL": "Israel",
	"IT": "Italy",
	"JM": "Jamaica",
	"JP": "Japan",
	"JE": "Jersey",
	"JO": "Jordan",
	"KZ": "Kazakhstan",
	"KE": "Kenya",
	"KI": "Kiribati",
	"KR": "Korea",
	"KP": "North Korea",
	"KW": "Kuwait",
	"KG": "Kyrgyzstan",
	"LA": "Lao People\"s Democratic Republic",
	"LV": "Latvia",
	"LB": "Lebanon",
	"LS": "Lesotho",
	"LR": "Liberia",
	"LY": "Libyan Arab Jamahiriya",
	"LI": "Liechtenstein",
	"LT": "Lithuania",
	"LU": "Luxembourg",
	"MO": "Macao",
	"MK": "Macedonia",
	"MG": "Madagascar",
	"MW": "Malawi",
	"MY": "Malaysia",
	"MV": "Maldives",
	"ML": "Mali",
	"MT": "Malta",
	"MH": "Marshall Islands",
	"MQ": "Martinique",
	"MR": "Mauritania",
	"MU": "Mauritius",
	"YT": "Mayotte",
	"MX": "Mexico",
	"FM": "Micronesia, Federated States Of",
	"MD": "Moldova",
	"MC": "Monaco",
	"MN": "Mongolia",
	"ME": "Montenegro",
	"MS": "Montserrat",
	"MA": "Morocco",
	"MZ": "Mozambique",
	"MM": "Myanmar",
	"NA": "Namibia",
	"NR": "Nauru",
	"NP": "Nepal",
	"NL": "Netherlands",
	"AN": "Netherlands Antilles",
	"NC": "New Caledonia",
	"NZ": "New Zealand",
	"NI": "Nicaragua",
	"NE": "Niger",
	"NG": "Nigeria",
	"NU": "Niue",
	"NF": "Norfolk Island",
	"MP": "Northern Mariana Islands",
	"NO": "Norway",
	"OM": "Oman",
	"PK": "Pakistan",
	"PW": "Palau",
	"PS": "Palestinian Territory, Occupied",
	"PA": "Panama",
	"PG": "Papua New Guinea",
	"PY": "Paraguay",
	"PE": "Peru",
	"PH": "Philippines",
	"PN": "Pitcairn",
	"PL": "Poland",
	"PT": "Portugal",
	"PR": "Puerto Rico",
	"QA": "Qatar",
	"RE": "Reunion",
	"RO": "Romania",
	"RU": "Russian Federation",
	"RW": "Rwanda",
	"BL": "Saint Barthelemy",
	"SH": "Saint Helena",
	"KN": "Saint Kitts And Nevis",
	"LC": "Saint Lucia",
	"MF": "Saint Martin",
	"PM": "Saint Pierre And Miquelon",
	"VC": "Saint Vincent And Grenadines",
	"WS": "Samoa",
	"SM": "San Marino",
	"ST": "Sao Tome And Principe",
	"SA": "Saudi Arabia",
	"SN": "Senegal",
	"RS": "Serbia",
	"SC": "Seychelles",
	"SL": "Sierra Leone",
	"SG": "Singapore",
	"SK": "Slovakia",
	"SI": "Slovenia",
	"SB": "Solomon Islands",
	"SO": "Somalia",
	"ZA": "South Africa",
	"GS": "South Georgia And Sandwich Isl.",
	"ES": "Spain",
	"LK": "Sri Lanka",
	"SD": "Sudan",
	"SR": "Suriname",
	"SJ": "Svalbard And Jan Mayen",
	"SZ": "Swaziland",
	"SE": "Sweden",
	"CH": "Switzerland",
	"SY": "Syrian Arab Republic",
	"TW": "Taiwan",
	"TJ": "Tajikistan",
	"TZ": "Tanzania",
	"TH": "Thailand",
	"TL": "Timor-Leste",
	"TG": "Togo",
	"TK": "Tokelau",
	"TO": "Tonga",
	"TT": "Trinidad And Tobago",
	"TN": "Tunisia",
	"TR": "Turkey",
	"TM": "Turkmenistan",
	"TC": "Turks And Caicos Islands",
	"TV": "Tuvalu",
	"UG": "Uganda",
	"UA": "Ukraine",
	"AE": "United Arab Emirates",
	"GB": "United Kingdom",
	"US": "United States",
	"UM": "United States Outlying Islands",
	"UY": "Uruguay",
	"UZ": "Uzbekistan",
	"VU": "Vanuatu",
	"VE": "Venezuela",
	"VN": "Vietnam",
	"VG": "Virgin Islands, British",
	"VI": "Virgin Islands, U.S.",
	"WF": "Wallis And Futuna",
	"EH": "Western Sahara",
	"YE": "Yemen",
	"ZM": "Zambia",
	"ZW": "Zimbabwe"
	}`
var countryList = map[string]string{}

func CountryIso2ToCountryName(iso2 string) string {
	if len(countryList) == 0 {
		json.Unmarshal([]byte(countryString), &countryList)
	}
	iso2 = strings.ToUpper(iso2)
	if name, ok := countryList[iso2]; ok {
		return name
	}
	return iso2
}

// Int64ToUint64Ptr convierte un int64 a *uint64 (negativos se consideran 0)
func Int64ToUint64Ptr(i int64) *uint64 {
	var u uint64
	if i < 0 {
		u = 0
	} else {
		u = uint64(i)
	}
	return &u
}
