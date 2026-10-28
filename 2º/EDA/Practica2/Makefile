CXX = g++
FLAGS = -Wall

all: costes graficas

costes: costes.cpp
	$(CXX) $(FLAGS) -o costes costes.cpp

graficas: costes
	./costes
	gnuplot graficas.gnuplot

clean:
	rm -f costes *.png *.dat
