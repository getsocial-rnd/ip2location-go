package ip2location

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/big"
	"net"
	"os"
	"strconv"
)

const (
	ApiVersion string = "8.0.3"

	countryshort       uint32 = 0x00001
	countrylong        uint32 = 0x00002
	region             uint32 = 0x00004
	city               uint32 = 0x00008
	isp                uint32 = 0x00010
	latitude           uint32 = 0x00020
	longitude          uint32 = 0x00040
	domain             uint32 = 0x00080
	zipcode            uint32 = 0x00100
	timezone           uint32 = 0x00200
	netspeed           uint32 = 0x00400
	iddcode            uint32 = 0x00800
	areacode           uint32 = 0x01000
	weatherstationcode uint32 = 0x02000
	weatherstationname uint32 = 0x04000
	mcc                uint32 = 0x08000
	mnc                uint32 = 0x10000
	mobilebrand        uint32 = 0x20000
	elevation          uint32 = 0x40000
	usagetype          uint32 = 0x80000

	all uint32 = countryshort | countrylong | region | city | isp | latitude | longitude | domain | zipcode | timezone | netspeed | iddcode | areacode | weatherstationcode | weatherstationname | mcc | mnc | mobilebrand | elevation | usagetype

	invalid_address string = "Invalid IP address."
	not_supported   string = "This parameter is unavailable for selected data file. Please upgrade the data file."
)

var (
	countryPosition            = [25]uint8{0, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2}
	regionPosition             = [25]uint8{0, 0, 0, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3}
	cityPosition               = [25]uint8{0, 0, 0, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4}
	ispPosition                = [25]uint8{0, 0, 3, 0, 5, 0, 7, 5, 7, 0, 8, 0, 9, 0, 9, 0, 9, 0, 9, 7, 9, 0, 9, 7, 9}
	latitudePosition           = [25]uint8{0, 0, 0, 0, 0, 5, 5, 0, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5}
	longitudePosition          = [25]uint8{0, 0, 0, 0, 0, 6, 6, 0, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6}
	domainPosition             = [25]uint8{0, 0, 0, 0, 0, 0, 0, 6, 8, 0, 9, 0, 10, 0, 10, 0, 10, 0, 10, 8, 10, 0, 10, 8, 10}
	zipCodePosition            = [25]uint8{0, 0, 0, 0, 0, 0, 0, 0, 0, 7, 7, 7, 7, 0, 7, 7, 7, 0, 7, 0, 7, 7, 7, 0, 7}
	timeZonePosition           = [25]uint8{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 8, 8, 7, 8, 8, 8, 7, 8, 0, 8, 8, 8, 0, 8}
	netSpeedPosition           = [25]uint8{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 8, 11, 0, 11, 8, 11, 0, 11, 0, 11, 0, 11}
	iddCodePosition            = [25]uint8{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 9, 12, 0, 12, 0, 12, 9, 12, 0, 12}
	areaCodePosition           = [25]uint8{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 13, 0, 13, 0, 13, 10, 13, 0, 13}
	weatherStationCodePosition = [25]uint8{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 9, 14, 0, 14, 0, 14, 0, 14}
	weatherStationNamePosition = [25]uint8{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 15, 0, 15, 0, 15, 0, 15}
	mccPosition                = [25]uint8{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 9, 16, 0, 16, 9, 16}
	mncPosition                = [25]uint8{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 17, 0, 17, 10, 17}
	mobileBrandPosition        = [25]uint8{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 11, 18, 0, 18, 11, 18}
	elevationPosition          = [25]uint8{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 11, 19, 0, 19}
	usageTypePosition          = [25]uint8{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 12, 20}
	maxIpv4Range               = big.NewInt(4294967295)
	maxIpv6Range               = big.NewInt(0)
)

type DB struct {
	file *os.File

	// DB specific offsets
	countryPositionOffset            uint32
	regionPositionOffset             uint32
	cityPositionOffset               uint32
	ispPositionOffset                uint32
	domainPositionOffset             uint32
	zipcodePositionOffset            uint32
	latitudePositionOffset           uint32
	longitudePositionOffset          uint32
	timeZonePositionOffset           uint32
	netSpeedPositionOffset           uint32
	iddCodePositionOffset            uint32
	areaCodePositionOffset           uint32
	weatherStationCodePositionOffset uint32
	weatherStationNamePositionOffset uint32
	mccPositionOffset                uint32
	mncPositionOffset                uint32
	mobileBrandPositionOffset        uint32
	elevationPositionOffset          uint32
	usageTypePositionOffset          uint32

	// Feature flags
	countryEnabled            bool
	regionEnabled             bool
	cityEnabled               bool
	ispEnabled                bool
	domainEnabled             bool
	zipCodeEnabled            bool
	latitudeEnabled           bool
	longitudeEnabled          bool
	timeZoneEnabled           bool
	netSpeedEnabled           bool
	iddCodeEnabled            bool
	areaCodeEnabled           bool
	weatherStationCodeEnabled bool
	weatherStationNameEnabled bool
	mccEnabled                bool
	mncEnabled                bool
	mobileBrandEnabled        bool
	elevationEnabled          bool
	usageTypeEnabled          bool

	meta *dbMeta
}

type dbMeta struct {
	databaseType      uint8
	databesColumn     uint8
	databaseDay       uint8
	databaseMonth     uint8
	databaseYear      uint8
	ipv4DatabaseCount uint32
	ipv4DatabaseAddr  uint32
	ipv6DatabaseCount uint32
	ipv6DatabaseAddr  uint32
	ipv4IndexBaseAddr uint32
	ipv6IndexBaseAddr uint32
	ipv4ColumnsSize   uint32
	ipv6ColumnSize    uint32
}

type Record struct {
	CountryShort       string
	CountryLong        string
	Region             string
	City               string
	Isp                string
	Latitude           float32
	Longitude          float32
	Domain             string
	Zipcode            string
	TimeZone           string
	NetSpeed           string
	IddCode            string
	Areacode           string
	WeatherStationCode string
	WeatherStationName string
	Mcc                string
	Mnc                string
	MobileBrand        string
	Elevation          float32
	UsageType          string
}

// Open opens the database file at the given path and initializes the database.
func Open(dbPath string) (*DB, error) {
	maxIpv6Range.SetString("340282366920938463463374607431768211455", 10)

	f, err := os.Open(dbPath)
	if err != nil {
		return nil, err
	}

	db := &DB{
		file: f,
		meta: &dbMeta{},
	}

	db.meta.databaseType = db.readUint8(1)
	db.meta.databesColumn = db.readUint8(2)
	db.meta.databaseYear = db.readUint8(3)
	db.meta.databaseMonth = db.readUint8(4)
	db.meta.databaseDay = db.readUint8(5)
	db.meta.ipv4DatabaseCount = db.readUint32(6)
	db.meta.ipv4DatabaseAddr = db.readUint32(10)
	db.meta.ipv6DatabaseCount = db.readUint32(14)
	db.meta.ipv6DatabaseAddr = db.readUint32(18)
	db.meta.ipv4IndexBaseAddr = db.readUint32(22)
	db.meta.ipv6IndexBaseAddr = db.readUint32(26)
	db.meta.ipv4ColumnsSize = uint32(db.meta.databesColumn << 2)             // 4 bytes each column
	db.meta.ipv6ColumnSize = uint32(16 + ((db.meta.databesColumn - 1) << 2)) // 4 bytes each column, except IPFrom column which is 16 bytes

	dbt := db.meta.databaseType

	// since both IPv4 and IPv6 use 4 bytes for the below columns, can just do it once here
	if countryPosition[dbt] != 0 {
		db.countryPositionOffset = uint32(countryPosition[dbt]-1) << 2
		db.countryEnabled = true
	}
	if regionPosition[dbt] != 0 {
		db.regionPositionOffset = uint32(regionPosition[dbt]-1) << 2
		db.regionEnabled = true
	}
	if cityPosition[dbt] != 0 {
		db.cityPositionOffset = uint32(cityPosition[dbt]-1) << 2
		db.cityEnabled = true
	}
	if ispPosition[dbt] != 0 {
		db.ispPositionOffset = uint32(ispPosition[dbt]-1) << 2
		db.ispEnabled = true
	}
	if domainPosition[dbt] != 0 {
		db.domainPositionOffset = uint32(domainPosition[dbt]-1) << 2
		db.domainEnabled = true
	}
	if zipCodePosition[dbt] != 0 {
		db.zipcodePositionOffset = uint32(zipCodePosition[dbt]-1) << 2
		db.zipCodeEnabled = true
	}
	if latitudePosition[dbt] != 0 {
		db.latitudePositionOffset = uint32(latitudePosition[dbt]-1) << 2
		db.latitudeEnabled = true
	}
	if longitudePosition[dbt] != 0 {
		db.longitudePositionOffset = uint32(longitudePosition[dbt]-1) << 2
		db.longitudeEnabled = true
	}
	if timeZonePosition[dbt] != 0 {
		db.timeZonePositionOffset = uint32(timeZonePosition[dbt]-1) << 2
		db.timeZoneEnabled = true
	}
	if netSpeedPosition[dbt] != 0 {
		db.netSpeedPositionOffset = uint32(netSpeedPosition[dbt]-1) << 2
		db.netSpeedEnabled = true
	}
	if iddCodePosition[dbt] != 0 {
		db.iddCodePositionOffset = uint32(iddCodePosition[dbt]-1) << 2
		db.iddCodeEnabled = true
	}
	if areaCodePosition[dbt] != 0 {
		db.areaCodePositionOffset = uint32(areaCodePosition[dbt]-1) << 2
		db.areaCodeEnabled = true
	}
	if weatherStationCodePosition[dbt] != 0 {
		db.weatherStationCodePositionOffset = uint32(weatherStationCodePosition[dbt]-1) << 2
		db.weatherStationCodeEnabled = true
	}
	if weatherStationNamePosition[dbt] != 0 {
		db.weatherStationNamePositionOffset = uint32(weatherStationNamePosition[dbt]-1) << 2
		db.weatherStationNameEnabled = true
	}
	if mccPosition[dbt] != 0 {
		db.mccPositionOffset = uint32(mccPosition[dbt]-1) << 2
		db.mccEnabled = true
	}
	if mncPosition[dbt] != 0 {
		db.mncPositionOffset = uint32(mncPosition[dbt]-1) << 2
		db.mncEnabled = true
	}
	if mobileBrandPosition[dbt] != 0 {
		db.mobileBrandPositionOffset = uint32(mobileBrandPosition[dbt]-1) << 2
		db.mobileBrandEnabled = true
	}
	if elevationPosition[dbt] != 0 {
		db.elevationPositionOffset = uint32(elevationPosition[dbt]-1) << 2
		db.elevationEnabled = true
	}
	if usageTypePosition[dbt] != 0 {
		db.usageTypePositionOffset = uint32(usageTypePosition[dbt]-1) << 2
		db.usageTypeEnabled = true
	}

	return db, nil
}

// Close closes the database.
func (db *DB) Close() error {
	return db.file.Close()
}

// get IP type and calculate IP number; calculates index too if exists
func (db *DB) checkIP(ip string) (iptype uint32, ipnum *big.Int, ipindex uint32) {
	iptype = 0
	ipnum = big.NewInt(0)
	ipnumtmp := big.NewInt(0)
	ipindex = 0
	ipaddress := net.ParseIP(ip)

	if ipaddress != nil {
		v4 := ipaddress.To4()

		if v4 != nil {
			iptype = 4
			ipnum.SetBytes(v4)
		} else {
			v6 := ipaddress.To16()

			if v6 != nil {
				iptype = 6
				ipnum.SetBytes(v6)
			}
		}
	}
	if iptype == 4 {
		if db.meta.ipv4IndexBaseAddr > 0 {
			ipnumtmp.Rsh(ipnum, 16)
			ipnumtmp.Lsh(ipnumtmp, 3)
			ipindex = uint32(ipnumtmp.Add(ipnumtmp, big.NewInt(int64(db.meta.ipv4IndexBaseAddr))).Uint64())
		}
	} else if iptype == 6 {
		if db.meta.ipv6IndexBaseAddr > 0 {
			ipnumtmp.Rsh(ipnum, 112)
			ipnumtmp.Lsh(ipnumtmp, 3)
			ipindex = uint32(ipnumtmp.Add(ipnumtmp, big.NewInt(int64(db.meta.ipv6IndexBaseAddr))).Uint64())
		}
	}
	return
}

// read byte
func (db *DB) readUint8(pos int64) uint8 {
	var retval uint8
	data := make([]byte, 1)
	_, err := db.file.ReadAt(data, pos-1)
	if err != nil {
		fmt.Println("File read failed:", err)
	}
	retval = data[0]
	return retval
}

// read unsigned 32-bit integer
func (db *DB) readUint32(pos uint32) uint32 {
	pos2 := int64(pos)
	var retval uint32
	data := make([]byte, 4)
	_, err := db.file.ReadAt(data, pos2-1)
	if err != nil {
		fmt.Println("File read failed:", err)
	}
	buf := bytes.NewReader(data)
	err = binary.Read(buf, binary.LittleEndian, &retval)
	if err != nil {
		fmt.Println("Binary read failed:", err)
	}
	return retval
}

// read unsigned 128-bit integer
func (db *DB) readUint128(pos uint32) *big.Int {
	pos2 := int64(pos)
	retval := big.NewInt(0)
	data := make([]byte, 16)
	_, err := db.file.ReadAt(data, pos2-1)
	if err != nil {
		fmt.Println("File read failed:", err)
	}

	// little endian to big endian
	for i, j := 0, len(data)-1; i < j; i, j = i+1, j-1 {
		data[i], data[j] = data[j], data[i]
	}
	retval.SetBytes(data)
	return retval
}

// read string
func (db *DB) readStr(pos uint32) string {
	pos2 := int64(pos)
	var retval string
	lenbyte := make([]byte, 1)
	_, err := db.file.ReadAt(lenbyte, pos2)
	if err != nil {
		fmt.Println("File read failed:", err)
	}
	strlen := lenbyte[0]
	data := make([]byte, strlen)
	_, err = db.file.ReadAt(data, pos2+1)
	if err != nil {
		fmt.Println("File read failed:", err)
	}
	retval = string(data[:strlen])
	return retval
}

// read float
func (db *DB) readFloat(pos uint32) float32 {
	pos2 := int64(pos)
	var retval float32
	data := make([]byte, 4)
	_, err := db.file.ReadAt(data, pos2-1)
	if err != nil {
		fmt.Println("File read failed:", err)
	}
	buf := bytes.NewReader(data)
	err = binary.Read(buf, binary.LittleEndian, &retval)
	if err != nil {
		fmt.Println("Binary read failed:", err)
	}
	return retval
}

// populate record with message
func (db *DB) loadMessage(mesg string) Record {
	var x Record

	x.CountryShort = mesg
	x.CountryLong = mesg
	x.Region = mesg
	x.City = mesg
	x.Isp = mesg
	x.Domain = mesg
	x.Zipcode = mesg
	x.TimeZone = mesg
	x.NetSpeed = mesg
	x.IddCode = mesg
	x.Areacode = mesg
	x.WeatherStationCode = mesg
	x.WeatherStationName = mesg
	x.Mcc = mesg
	x.Mnc = mesg
	x.MobileBrand = mesg
	x.UsageType = mesg

	return x
}

// get all fields
func (db *DB) GetAll(ipaddress string) Record {
	return db.query(ipaddress, all)
}

// get country code
func (db *DB) GetCountryShort(ipaddress string) Record {
	return db.query(ipaddress, countryshort)
}

// get country name
func (db *DB) GetCountryLong(ipaddress string) Record {
	return db.query(ipaddress, countrylong)
}

// get region
func (db *DB) GetRegion(ipaddress string) Record {
	return db.query(ipaddress, region)
}

// get city
func (db *DB) GetCity(ipaddress string) Record {
	return db.query(ipaddress, city)
}

// get isp
func (db *DB) GetISP(ipaddress string) Record {
	return db.query(ipaddress, isp)
}

// get latitude
func (db *DB) GetLatitude(ipaddress string) Record {
	return db.query(ipaddress, latitude)
}

// get longitude
func (db *DB) GetLongitude(ipaddress string) Record {
	return db.query(ipaddress, longitude)
}

// get domain
func (db *DB) GetDomain(ipaddress string) Record {
	return db.query(ipaddress, domain)
}

// get zip code
func (db *DB) GetZipCode(ipaddress string) Record {
	return db.query(ipaddress, zipcode)
}

// get time zone
func (db *DB) GetTimeZone(ipaddress string) Record {
	return db.query(ipaddress, timezone)
}

// get net speed
func (db *DB) GetNetSpeed(ipaddress string) Record {
	return db.query(ipaddress, netspeed)
}

// get idd code
func (db *DB) GetIDDCode(ipaddress string) Record {
	return db.query(ipaddress, iddcode)
}

// get area code
func (db *DB) GetAreaCode(ipaddress string) Record {
	return db.query(ipaddress, areacode)
}

// get weather station code
func (db *DB) GetWeatherStationCode(ipaddress string) Record {
	return db.query(ipaddress, weatherstationcode)
}

// get weather station name
func (db *DB) GetWeatherStationName(ipaddress string) Record {
	return db.query(ipaddress, weatherstationname)
}

// get mobile country code
func (db *DB) GetMCC(ipaddress string) Record {
	return db.query(ipaddress, mcc)
}

// get mobile network code
func (db *DB) GetMNC(ipaddress string) Record {
	return db.query(ipaddress, mnc)
}

// get mobile carrier brand
func (db *DB) GetMobileBrand(ipaddress string) Record {
	return db.query(ipaddress, mobilebrand)
}

// get elevation
func (db *DB) GetElevation(ipaddress string) Record {
	return db.query(ipaddress, elevation)
}

// get usage type
func (db *DB) GetUsageType(ipaddress string) Record {
	return db.query(ipaddress, usagetype)
}

// main query
func (db *DB) query(ipaddress string, mode uint32) Record {
	x := db.loadMessage(not_supported) // default message

	// check IP type and return IP number & index (if exists)
	iptype, ipno, ipindex := db.checkIP(ipaddress)

	if iptype == 0 {
		x = db.loadMessage(invalid_address)
		return x
	}

	var colsize uint32
	var baseaddr uint32
	var low uint32
	var high uint32
	var mid uint32
	var rowoffset uint32
	var rowoffset2 uint32
	ipfrom := big.NewInt(0)
	ipto := big.NewInt(0)
	maxip := big.NewInt(0)

	if iptype == 4 {
		baseaddr = db.meta.ipv4DatabaseAddr
		high = db.meta.ipv4DatabaseCount
		maxip = maxIpv4Range
		colsize = db.meta.ipv4ColumnsSize
	} else {
		baseaddr = db.meta.ipv6DatabaseAddr
		high = db.meta.ipv6DatabaseCount
		maxip = maxIpv6Range
		colsize = db.meta.ipv6ColumnSize
	}

	// reading index
	if ipindex > 0 {
		low = db.readUint32(ipindex)
		high = db.readUint32(ipindex + 4)
	}

	if ipno.Cmp(maxip) >= 0 {
		ipno = ipno.Sub(ipno, big.NewInt(1))
	}

	for low <= high {
		mid = (low + high) >> 1
		rowoffset = baseaddr + (mid * colsize)
		rowoffset2 = rowoffset + colsize

		if iptype == 4 {
			ipfrom = big.NewInt(int64(db.readUint32(rowoffset)))
			ipto = big.NewInt(int64(db.readUint32(rowoffset2)))
		} else {
			ipfrom = db.readUint128(rowoffset)
			ipto = db.readUint128(rowoffset2)
		}

		if ipno.Cmp(ipfrom) >= 0 && ipno.Cmp(ipto) < 0 {
			if iptype == 6 {
				rowoffset = rowoffset + 12 // coz below is assuming all columns are 4 bytes, so got 12 left to go to make 16 bytes total
			}

			if mode&countryshort == 1 && db.countryEnabled {
				x.CountryShort = db.readStr(db.readUint32(rowoffset + db.countryPositionOffset))
			}

			if mode&countrylong != 0 && db.countryEnabled {
				x.CountryLong = db.readStr(db.readUint32(rowoffset+db.countryPositionOffset) + 3)
			}

			if mode&region != 0 && db.regionEnabled {
				x.Region = db.readStr(db.readUint32(rowoffset + db.regionPositionOffset))
			}

			if mode&city != 0 && db.cityEnabled {
				x.City = db.readStr(db.readUint32(rowoffset + db.cityPositionOffset))
			}

			if mode&isp != 0 && db.ispEnabled {
				x.Isp = db.readStr(db.readUint32(rowoffset + db.ispPositionOffset))
			}

			if mode&latitude != 0 && db.latitudeEnabled {
				x.Latitude = db.readFloat(rowoffset + db.latitudePositionOffset)
			}

			if mode&longitude != 0 && db.longitudeEnabled {
				x.Longitude = db.readFloat(rowoffset + db.longitudePositionOffset)
			}

			if mode&domain != 0 && db.domainEnabled {
				x.Domain = db.readStr(db.readUint32(rowoffset + db.domainPositionOffset))
			}

			if mode&zipcode != 0 && db.zipCodeEnabled {
				x.Zipcode = db.readStr(db.readUint32(rowoffset + db.zipcodePositionOffset))
			}

			if mode&timezone != 0 && db.timeZoneEnabled {
				x.TimeZone = db.readStr(db.readUint32(rowoffset + db.timeZonePositionOffset))
			}

			if mode&netspeed != 0 && db.netSpeedEnabled {
				x.NetSpeed = db.readStr(db.readUint32(rowoffset + db.netSpeedPositionOffset))
			}

			if mode&iddcode != 0 && db.iddCodeEnabled {
				x.IddCode = db.readStr(db.readUint32(rowoffset + db.iddCodePositionOffset))
			}

			if mode&areacode != 0 && db.areaCodeEnabled {
				x.Areacode = db.readStr(db.readUint32(rowoffset + db.areaCodePositionOffset))
			}

			if mode&weatherstationcode != 0 && db.weatherStationCodeEnabled {
				x.WeatherStationCode = db.readStr(db.readUint32(rowoffset + db.weatherStationCodePositionOffset))
			}

			if mode&weatherstationname != 0 && db.weatherStationNameEnabled {
				x.WeatherStationName = db.readStr(db.readUint32(rowoffset + db.weatherStationNamePositionOffset))
			}

			if mode&mcc != 0 && db.mccEnabled {
				x.Mcc = db.readStr(db.readUint32(rowoffset + db.mccPositionOffset))
			}

			if mode&mnc != 0 && db.mncEnabled {
				x.Mnc = db.readStr(db.readUint32(rowoffset + db.mncPositionOffset))
			}

			if mode&mobilebrand != 0 && db.mobileBrandEnabled {
				x.MobileBrand = db.readStr(db.readUint32(rowoffset + db.mobileBrandPositionOffset))
			}

			if mode&elevation != 0 && db.elevationEnabled {
				f, _ := strconv.ParseFloat(db.readStr(db.readUint32(rowoffset+db.elevationPositionOffset)), 32)
				x.Elevation = float32(f)
			}

			if mode&usagetype != 0 && db.usageTypeEnabled {
				x.UsageType = db.readStr(db.readUint32(rowoffset + db.usageTypePositionOffset))
			}

			return x
		} else {
			if ipno.Cmp(ipfrom) < 0 {
				high = mid - 1
			} else {
				low = mid + 1
			}
		}
	}
	return x
}

func (x Record) String() string {
	buf := &bytes.Buffer{}
	fmt.Fprintf(buf, "country_short: %s\n", x.CountryShort)
	fmt.Fprintf(buf, "country_long: %s\n", x.CountryLong)
	fmt.Fprintf(buf, "region: %s\n", x.Region)
	fmt.Fprintf(buf, "city: %s\n", x.City)
	fmt.Fprintf(buf, "isp: %s\n", x.Isp)
	fmt.Fprintf(buf, "latitude: %file\n", x.Latitude)
	fmt.Fprintf(buf, "longitude: %file\n", x.Longitude)
	fmt.Fprintf(buf, "domain: %s\n", x.Domain)
	fmt.Fprintf(buf, "zipcode: %s\n", x.Zipcode)
	fmt.Fprintf(buf, "timezone: %s\n", x.TimeZone)
	fmt.Fprintf(buf, "netspeed: %s\n", x.NetSpeed)
	fmt.Fprintf(buf, "iddcode: %s\n", x.IddCode)
	fmt.Fprintf(buf, "areacode: %s\n", x.Areacode)
	fmt.Fprintf(buf, "weatherstationcode: %s\n", x.WeatherStationCode)
	fmt.Fprintf(buf, "weatherstationname: %s\n", x.WeatherStationName)
	fmt.Fprintf(buf, "mcc: %s\n", x.Mcc)
	fmt.Fprintf(buf, "mnc: %s\n", x.Mnc)
	fmt.Fprintf(buf, "mobilebrand: %s\n", x.MobileBrand)
	fmt.Fprintf(buf, "elevation: %file\n", x.Elevation)
	fmt.Fprintf(buf, "usagetype: %s\n", x.UsageType)
	return buf.String()
}
