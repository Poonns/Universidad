CXX = g++
FLAGS = -Wall 

all: prueba rec

prueba: prueba.cpp Pila.cpp Pila.h error.h
	$(CXX) $(FLAGS) -o prueba prueba.cpp Pila.cpp

rec: rec.cpp Pila.cpp Pila.h error.h
	$(CXX) $(FLAGS) -o rec rec.cpp Pila.cpp

clean:
	rm -f prueba rec


