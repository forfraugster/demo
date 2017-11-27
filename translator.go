package main

import (
	"fmt"
	"log"
	//"strconv"
	"os"
	"encoding/binary"
	"bytes"
	"github.com/hydrogen18/stalecucumber"
	"flag"
	/*"bufio"
	"io/ioutil"*/
	"reflect"
)

type Header struct { /*todo-ad05bzag what if i type out the header as one massive slice and then just keep on chopping it?*/
	//StartHeaderA																		[4]byte //35,35 empty empty dec 23,23,0,0 hex
	FirstTimestamp, LastTimestamp                                                       [4]byte
	LineCount, SamplingRate, IdentReactor                                               [4]byte
	Version, Threshold, BatchState, NumberInput, NumberControlLoop, NumberOutputChannel [1]uint8
	TagXCUBIO                                                                           [41]byte //todo-ad05bzag check the count here. 40 or 41?!
	Comment                                                                             [253]byte
	CRCCheckA                                                                           [4]byte
	EndHeaderA                                                                          [4]uint8 //35,35,13,10 dec 23,23,0D,0A hex
	StartHeaderB                                                                        [4]uint8 //40,40,40,40 dec 28,28,28,28 hex
	TrendChannelNames                                                                   [1375]byte
	MidHeaderB                                                                          [4]uint8 //40,40,13,10 dec 28,28,0D,0A hex
	TrendUnitNames                                                                      [1377]byte
	CRCCheckB                                                                           [4]uint8
	EndHeaderB                                                                          [4]uint8 //40,40,13,10 dec 28,28,0D,0A hex
	StartData                                                                           [4]uint8
	CurrentTimestamp                                                                    [4]uint8
	NumberInputChanged                                                                  [4]uint8
	NumberControlLoopChanged                                                            [4]uint8
	NumberOutputChannelChanged                                                          [4]uint8
	Rest                                                                                [5000]byte // this a placeholder for the actual business data
}

var screenPrint, pickle *bool
var trend string

func init() {
	pickle = flag.Bool("pickle", false, "you wanna pickle dat badboi?")       //https://stackoverflow.com/questions/27411691/how-to-pass-boolean-arguments-to-go-flags
	screenPrint = flag.Bool("print", false, "you wanna print in on screen?")
	flag.StringVar(&trend, "trend", "", "this will be the .TREND to work on") //https://gobyexample.com/command-line-flags
	flag.Parse()
}

func check(e error) { //supposedly that makes it easier to check calls for errors). But as per effective go, panic for errors is not a good idea. todo-ad05bzag modify to NOT throw panic()
	if e != nil {
		panic(e)
	}
}

func readNextBytes(file *os.File, number int) []byte {
	b := make([]byte, number)
	_, err := file.Read(b)
	check(err)
	return b
}

func fileExists(n string) bool { //checks whether file exists and returns a bool value
	if _, err := os.Stat(n); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func (f *Header) reflect() { //for reference https://blog.golang.org/laws-of-reflection and https://gist.github.com/drewolson/4771479 todo-ad05bzag better ERROR HANDLING for reflection.
	val := reflect.ValueOf(f).Elem()
	for i := 0; i < val.NumField(); i++ {
		valueField := val.Field(i)
		typeField := val.Type().Field(i)
		typeDecode := func(i interface{}) {
			switch i.(type) {
			case [1]uint8:
				fmt.Printf("Field Name: %-30s|\t Actual value: %d\n", typeField.Name, valueField.Interface())
				//noinspection ALL
			case [4]uint8: //todo-ad05bzag Add maths/bit for func TrailingZeros8 (x uint8) int, such that -> 8 when x == 0
				if typeField.Name == "FirstTimestamp" || typeField.Name == "LastTimestamp" || typeField.Name == "LastDat" || typeField.Name == "LineCount" || typeField.
					Name == "SamplingRate" || typeField.Name == "IdentReactor" || typeField.Name == "CurrentTimestamp" {
					var end []byte
					k := valueField.Interface().([4]byte)

					end = k[:]
					data := binary.LittleEndian.Uint32(end)

					fmt.Printf("Field Name: %-30s|\t Actual Value: %d\n", typeField.Name, data)
				} else {
					fmt.Printf("Field Name: %-30s|\t Test Byte Value: %b\n", typeField.Name, valueField.Interface())
				}
			case [41]uint8:
				fmt.Sprintf("Field Name: %-30s|\t Test Byte Value: %v\n", typeField.Name, valueField.Interface())
			case [253]uint8:
				fmt.Printf("Field Name: %-30s|\t Actual value: %+s\n", typeField.Name, valueField.Interface())
			case [1375]uint8:
				fmt.Printf("Field Name: %-30s|\t Actual value: %+q\n", typeField.Name, valueField.Interface())
			case [1377]uint8:
				fmt.Printf("Field Name: %-30s|\t Actual value: %+q\n", typeField.Name, valueField.Interface())
			default:
				fmt.Printf("Field Name: %-30s|\t Byte value: %#v\n", typeField.Name, valueField.Interface())
			}
		}
		typeDecode(valueField.Interface())
	}
}

func main() { //the strconv.Quote() ie converting to string was ugly af. upd: create byteslice
	//and then use bytes.Equal(a,b []byte) bool
	defer os.Exit(0) //closes all goroutines and exists gracefully?

	forRealB := []byte{0x23, 0x23, 0x0, 0x0}
	fmt.Printf("Expected first 4 bytes are:%#v\n", forRealB)

	path := trend
	file, err := os.Open(path)
	check(err)
	fmt.Printf("Working on this file: %s\n", path)
	defer file.Close() //till the block is done the function will not close the file

	delimit := readNextBytes(file, 4)
	fmt.Printf("%s is open. First 4 bytes are:%#v\n", path, delimit)

	header := Header{}
	if bytes.Equal(forRealB, delimit) {
		fmt.Printf("%s is an xCUBIO file\n", path)
		data := readNextBytes(file, 10000) //here indicates the number of bytes to be read. And i never know how much to allocate (c) todo-ad05bzag make the number dependent on what - size of buffer, some arbitrary system/network related capacity?
		buffer := bytes.NewBuffer(data)
		err = binary.Read(buffer, binary.LittleEndian, &header)
		check(err)
		if *screenPrint == true {
			header.reflect()
		}
	} else {
		log.Fatalf("%s is NOT an xCUBIO file\n", path)
	}

	if *pickle == true {
		fmt.Println("you says you wanna be pickling!? Hold, need to run a quick check")
		var name string
		name = path + ".binary"
		if !fileExists(name) {
			buf := new(bytes.Buffer)
			//fmt.Printf("here it is: %v\n, %v\n", buf, err)
			f, err := stalecucumber.NewPickler(buf).Pickle(header) // here f is the number of bytes written from the datastream, whereas buf has type of {*bytes.Buffer} and is the new pickled binary
			check(err)
			fmt.Printf("Number of bytes pickled: %d\n", f)
			outf, err := os.Create(name)
			check(err)
			defer outf.Close()
			f1, err := buf.WriteTo(outf)
			check(err)
			fmt.Printf("Number of bytes written: %d\n", f1)
			f_int64 := int64(f)
			f1_int64 := int64(f1)
			if f_int64 == f1_int64 {
				fmt.Println("we wrote the same number of bytes as we pickled")

			} else {
				log.Fatal("some shit went wrong and we wrote different number than we pickled")
			}
		} else {
			log.Fatalf("%s already exists. Aborting writing to pickle file!\n", name)
		}
	}
}
