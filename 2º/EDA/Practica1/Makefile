obj = complejo.o complejoMain.o
CC = g++
WARNING = -Wall

all: complejoMain

complejo.o: complejo.cpp complejo.h
	$(CC) -c complejo.cpp $(WARNING)

complejoMain.o: complejoMain.cpp complejo.h
	$(CC) -c complejoMain.cpp $(WARNING)

complejoMain: $(obj)
	$(CC) -o complejoMain $(obj)

clean:
	rm -f *.o *.gch complejoMain
