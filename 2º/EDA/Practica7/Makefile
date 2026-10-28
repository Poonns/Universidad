CC = g++
CCFLAGS = -Wall -std=c++11
OBJ_TEST = test_prim.o Grafo.o
OBJ_COSTES = costes.o Grafo.o

all: test_prim costes

test_prim: $(OBJ_TEST)
	$(CC) $(CCFLAGS) -o test_prim $(OBJ_TEST)

costes: $(OBJ_COSTES)
	$(CC) $(CCFLAGS) -o costes $(OBJ_COSTES)

Grafo.o: Grafo.cpp Grafo.h
	$(CC) $(CCFLAGS) -c Grafo.cpp

test_prim.o: test_prim.cpp Grafo.h Prim.h
	$(CC) $(CCFLAGS) -c test_prim.cpp

costes.o: costes.cpp Grafo.h Prim.h
	$(CC) $(CCFLAGS) -c costes.cpp

clean:
	rm -f *.o test_prim costes costes.dat script.gp costes.png