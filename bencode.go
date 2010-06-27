package bencode

import "strconv"
import "bytes"
import "fmt"
import "container/list"
import "os"
import "bufio"
/*
string: number:chars
int: i numbers e
dict: d string:|int|list|dict|string e
list: l string|int|list|dicte

*/
const (
	Bestr	= iota;
	Beint;
	Bedict;
	Belist;
	Beend;
)
/*
Just a way to split them up, 

*/

func NextToken(input *bufio.Reader)(encoded []byte, betype int, err os.Error){
	/* 
		dict returns just d
		list returns just l
		string returns encoded string
		int returns encoded int

	*/
	
	c, err := input.ReadByte();
	if err == os.EOF {
		return nil, 0,err
	}
	if err != nil {
		//print(err.String());
		return nil, 0,err;
	}

	switch {
	case c == 'i':	// it's an integer i..e
		//i:= 0;
		number := make([]byte,1,10);
		c, err = input.ReadByte();
		if err != nil {
			return nil, 0,err
		}
		for c != 'e' {
			number = bytes.AddByte(number, c);
			if c, err = input.ReadByte(); err != nil {
				return;
			}
		}
		return number,Beint,nil;

	case c == 'l':	//it's a list l...e
		alist := make([]byte,1,1);
		alist = bytes.AddByte(alist,c);
		return alist,Belist,nil;

	case c == 'e':	// the end of a dict or list
		end := make([]byte,1,1);
		end = bytes.AddByte(end,c);
		return end,Beend,nil

	case c >= '0' && c <= '9':	// it's a string and c is the size
		str := make([]byte,1,10);
		strSize := make([]byte,1,10);
		strSize = bytes.AddByte(strSize, c);	
		str = bytes.AddByte(str, c);
		var c1 uint8;
		c1, err  = input.ReadByte();
		if err != nil {
			return nil, 0,err
		}
		for c1 >= '0' && c1 <= '9' {
			strSize = bytes.AddByte(strSize, c1);	
			str= bytes.AddByte(str, c1);

			c1, err = input.ReadByte();
			if err != nil {
				return nil, 0, err
			}
		}
		strLen, err := strconv.Atoi(string(strSize));
		if err != nil {
			return nil, 0, os.NewError("strlength doesn't convert to int")
		}
		for i := 0; i < strLen; i++ {
			c2, err := input.ReadByte();
			if err != nil {
				return nil, 0, err
			}
			//fmt.Printf("the c is: %s\n",string(c1));
			str = bytes.AddByte(str, c2);
		}
		return str,Bestr,nil;

	case c == 'd':	//it's a dict
		dict := make( []byte,1,1);
		dict = bytes.AddByte(dict,c);
		return dict,Bedict,nil;
	}
	//shouldn't get here
	return nil,0,os.NewError("error, shouldn't be here");
}

type BeString []byte

type BeNode struct {
	Betype	int;
	Bestr	string;
	Beint	int;
	Bedict	map[string]*BeNode;
	Belist	*list.List;
}


/*finds the info dict and returns a hash of it*/
	/*infohash := sha1.New();Write(strings.Bytes(str));
	encode each item in be.Bedict["info"] and write to infohash
	
	be.Bedict["info_hash"] = infohash.Sum();	
	thing := sha1.New();
	str := "pizza";
	thing.Write(strings.Bytes(str));*/


func (this *BeNode) Encode() (output string, err os.Error) {
	buffer := new(bytes.Buffer);

	switch this.Betype {
	case Bestr:
		if _, err = buffer.WriteString(strconv.Itoa(len(this.Bestr))); err != nil {
			return
		}
		if _, err = buffer.WriteString(":"); err != nil {
			return
		}
		if _, err = buffer.WriteString(this.Bestr); err != nil {
			return
		}
	case Beint:
		if _, err = buffer.WriteString("i"); err != nil {
			return
		}
		if _, err = buffer.WriteString(strconv.Itoa(this.Beint)); err != nil {
			return
		}
		if _, err = buffer.WriteString("e"); err != nil {
			return
		}
	case Bedict:
		if _, err = buffer.WriteString("d"); err != nil {
			return
		}
		for key, item := range this.Bedict {
			var encoded string;
			if _, err = buffer.WriteString(strconv.Itoa(len(key))); err != nil {
				return
			}
			if _, err = buffer.WriteString(":"); err != nil {
				return
			}
			if _, err := buffer.WriteString(key); err != nil {
				return
			}
			if encoded, err = item.Encode(); err != nil {
				return
			}
			if _, err = buffer.WriteString(encoded); err != nil {
				return
			}
		}
		if _, err = buffer.WriteString("e"); err != nil {
			return
		}
	case Belist:

		if _, err = buffer.WriteString("l"); err != nil {
			return
		}
		for item := range this.Belist.Iter() {
			var encoded string;
			if encoded, err = item.(*BeNode).Encode(); err != nil {
				return
			}
			if _, err := buffer.WriteString(encoded); err != nil {
				return
			}
		}
		if _, err = buffer.WriteString("e"); err != nil {
			return
		}
	}
	output = buffer.String();
	return;
}

func (this *BeNode) Print() {
	switch this.Betype {
	case Bestr:
		//print("this is string:");
		fmt.Printf("%s\n", this.Bestr)
	case Beint:
		fmt.Printf("%d\n", this.Beint)
	case Belist:
		listchan := this.Belist.Iter();
		for i := range (listchan) {
			i.(*BeNode).Print()
		}
	case Bedict:
		//print("this is a dict");
		for i, j := range this.Bedict {
			fmt.Printf("%s=", i);
			j.Print();
			fmt.Printf("\n");
		}
	default:
		fmt.Printf("error")
	}
	return;
}

