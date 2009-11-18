package main
/* it's all text */
import "strings"
import "strconv"
import "bytes"
import "fmt"
import "container/list"

const(bestr=iota;beint;bedict;belist)

type BeString []byte

type BeNode struct {
	betype int;
	bestr string;
	beint int;
	bedict map[string] BeNode;
	belist list.List;
}

func (*BeString) AppendString(toencode string) []byte{
	str := strings.Bytes(strconv.Itoa(len(toencode)));
	result := make([]byte, len(toencode)+len(str)+1);
	result = bytes.Add(result,str);
	result = bytes.AddByte(result,':');
	result = bytes.Add(result,strings.Bytes(toencode));
	return result;
}

/*
needs to be iterable

*/
func (*BeString) DecodeString(input []byte) BeNode {

}

func (*BeString) DecodeInteger(input []byte) BeNode {

}

func (this *BeString) Decode(input []byte) BeNode {
	c,size:= utf8.DecodeRune(input);
	if(unicode.IsDigit(c)){
		switch c {
			case c =='i': // it's an integer
			case c == 'l': //it's a list

			case unicode.IsDigit(c): // it's a string

		}
	}
	
	split := bytes.Split(todecode,":", 0);
	length := atoi(split[0]);
	retString := string(split[1])
	

}

/*func decode(dst []byte, src []byte)(n int,typeof int, error Error){
	for i,dst[0] := range src {
		switch dst[0] {
			case 'i': // it's an integer
		}
	}
}*/

func (*BeString) AppendInteger(toencode int) []byte {
	str := strconv.Itoa(toencode);
	result := make([]byte,len(str)+2);
	result = bytes.AddByte(result,'i');
	result = bytes.Add(result,strings.Bytes(str));
	result = bytes.AddByte(result,'e');
	return result;
}

/*func EncodeList(toencode *list) []byte {
	
}*/

func main(){
	buff := new(BeString);
	buff.AppendString("pizza");
	buff.AppendInteger(-800);
	fmt.Printf("\n");
}
