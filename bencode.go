package main

import "strings"
import "strconv"
import "bytes"
import "fmt"
import "container/list"
import "os"
import "bufio"

const (
	bestr	= iota;
	beint;
	bedict;
	belist;
)

type BeString []byte

type BeNode struct {
	Betype	int;
	Bestr	string;
	Beint	int;
	Bedict	map[string]*BeNode;
	Belist	*list.List;
}

func (this *BeNode) Encode() (output string, err os.Error) {
	buffer := new(bytes.Buffer);

	switch this.Betype {
	case bestr:
		if _, err = buffer.WriteString(strconv.Itoa(len(this.Bestr))); err != nil {
			return
		}
		if _, err = buffer.WriteString(":"); err != nil {
			return
		}
		if _, err = buffer.WriteString(this.Bestr); err != nil {
			return
		}
	case beint:
		if _, err = buffer.WriteString("i"); err != nil {
			return
		}
		if _, err = buffer.WriteString(strconv.Itoa(this.Beint)); err != nil {
			return
		}
		if _, err = buffer.WriteString("e"); err != nil {
			return
		}
	case bedict:
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
	case belist:

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


func (this *BeString) Decode(input *bufio.Reader) (*BeNode, os.Error) {
	c, err := input.ReadByte();
	if err == os.EOF {
		return nil, err
	}
	if err != nil {
		print(err.String());
		return nil, err;
	}
	switch {
	case c == 'i':	// it's an integer i..e
		//i:= 0;
		var number []byte;
		c, err = input.ReadByte();
		if err != nil {
			return nil, err
		}
		for c != 'e' {
			number = bytes.AddByte(number, c);
			c, err = input.ReadByte();
		}
		node := new(BeNode);
		node.Betype = beint;
		node.Beint, err = strconv.Atoi(string(number));
		if err != nil {
			return nil, os.NewError("int doesn't convert to int")
		}
		return node, nil;

	case c == 'l':	//it's a list l...e
		node := new(BeNode);
		node.Betype = belist;
		node.Belist = list.New();
		itemnode, err := this.Decode(input);
		for ; itemnode != nil; itemnode, err = this.Decode(input) {
			node.Belist.PushBack(itemnode)
		}

		if itemnode == nil && err == nil {
			//end of list
			return node, nil
		} else {
			return nil, os.NewError("error in list item")
		}
	case c == 'e':	// the end of a dict or list
		return nil, nil

	case c > 47 && c < 58:	// it's a string and c is the size
		var str []byte;
		var strSize []byte;
		strSize = bytes.AddByte(strSize, c);
		c1, err := input.ReadByte();
		if err != nil {
			return nil, err
		}
		for c1 > 47 && c1 < 58 {
			strSize = bytes.AddByte(strSize, c1);
			c1, err = input.ReadByte();
			if err != nil {
				return nil, err
			}
		}
		strLen, err := strconv.Atoi(string(strSize));
		if err != nil {
			return nil, os.NewError("strlength doesn't convert to int")
		}
		for i := 0; i < strLen; i++ {
			c2, err := input.ReadByte();
			if err != nil {
				return nil, err
			}
			//fmt.Printf("the c is: %s\n",string(c1));
			str = bytes.AddByte(str, c2);
		}
		node := new(BeNode);
		node.Betype = bestr;
		node.Bestr = string(str);
		return node, nil;
	case c == 'd':	//it's a dict d...e
		node := new(BeNode);
		node.Betype = bedict;
		node.Bedict = make(map[string]*BeNode);
		keynode, keyerr := this.Decode(input);
		for ; keynode != nil; keynode, keyerr = this.Decode(input) {
			if keyerr == os.EOF {
				return node, nil
			}
			if keynode.Betype == bestr {
				//fmt.Printf("found key: %s\n",keynode.Bestr);
				itemnode, itemerr := this.Decode(input);
				if itemerr == nil && itemnode == nil {
					return node, nil
				}
				if itemerr == os.EOF {
					return node, nil
				}
				if itemnode == nil {
					return nil, itemerr
				}
				//itemnode.Print();
				node.Bedict[keynode.Bestr] = itemnode;

			} else {
				return nil, os.NewError("non string key in dict")
			}
		}

		if keynode == nil && keyerr == nil {
			//end of list
			return node, nil
		}

		return node, nil;

	}
	return nil, os.NewError("");
	;
}

func (this *BeNode) Print() {
	switch this.Betype {
	case bestr:
		//print("this is string:");
		fmt.Printf("%s\n", this.Bestr)
	case beint:
		fmt.Printf("%d\n", this.Beint)
	case belist:
		listchan := this.Belist.Iter();
		for i := range (listchan) {
			i.(*BeNode).Print()
		}
	case bedict:
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

func main() {
	buff := new(BeString);
	reader := bufio.NewReader(os.Stdin);
	be, err := buff.Decode(reader);
	if err != nil {
		fmt.Printf("%s\n", err.String());
		os.Exit(1);
	}
	//be.Print();
	be1 := new(BeNode);
	be1.Betype = bestr;
	be1.Bestr = "pizza";
	str1, err := be1.Encode();
	//print(str1);
	//print("\n");
	_ = str1;

	be2 := new(BeNode);
	be2.Betype = beint;
	be2.Beint = 500;
	str2, err := be2.Encode();
	//print(str2);
	//print("\n");
	_ = str2;

	be3 := new(BeNode);
	be3.Betype = belist;
	be3.Belist = new(list.List);

	be3.Belist.PushBack(be1);
	be3.Belist.PushBack(be2);

	str3, err := be3.Encode();
	//print(str3);
	//print("\n");
	_ = str3;

	be4 := new(BeNode);
	be4.Betype = bedict;
	be4.Bedict = make(map[string]*BeNode);
	//be4.Bedict["chickens"] = be1;
	//be4.Bedict["pie"] = be2;
	//be4.Bedict["cheese"] = be3;
	str4, err := be4.Encode();
	//	print(str4);
	//print("\n");
	str5, err := be.Encode();
	_ = err;
	_ = str4;
	print(str5);
}
