GL=8l
GC=8g

all: gobit.8
	${GL} -o gobit gobit.8
test_gotracker: gotracker.8 test_gotracker.8
	${GL} -o test_gotracker gotracker.8 test_gotracker.8

test_gotorrent: gotorrent.8
	${GL} -o test_gotorrent gotorrent.8

gobit: gobit.8
gotorrent: gotorrent.8
bencode: bencode.8
gotracker: gotracker.8

gobit.8: gotorrent.8 gotracker.8
	${GC} gobit.go

gotorrent.8: bencode.8
	${GC} gotorrent.go

gotracker.8: bencode.8
	${GC} gotracker.go

bencode.8: 
	${GC} bencode.go

test_gotracker.8:
	${GC} test_gotracker.go
clean:
	rm *.8