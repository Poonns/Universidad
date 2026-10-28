CC = g++
CFLAGS = -Wall -std=c++11 -g

# Reglas de compilación de objetos

alumno.o: alumno.cpp alumno.h
	$(CC) $(CFLAGS) -c alumno.cpp

test_arbol.o: test_arbol.cpp alumno.h abb.h
	$(CC) $(CFLAGS) -c test_arbol.cpp

test_map.o: test_map.cpp alumno.h
	$(CC) $(CFLAGS) -c test_map.cpp

# Reglas de enlazado y ejecutables

all: run_arbol test_map

run_arbol: test_arbol.o alumno.o
	$(CC) $(CFLAGS) -o run_arbol test_arbol.o alumno.o

test_map: test_map.o alumno.o
	$(CC) $(CFLAGS) -o test_map test_map.o alumno.o

# Reglas auxiliares

.PHONY: clean run_arbol run_map

clean:
	rm -f *.o run_arbol test_map

run_arbol: run_arbol
	./run_arbol

run_map: test_map
	./test_map
