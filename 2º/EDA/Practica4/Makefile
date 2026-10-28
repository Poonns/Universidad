CXX = g++
FLAGS = -Wall -O2

all: merge

merge: merge.cpp VElement.cpp VElement.h
	$(CXX) $(FLAGS) -o merge merge.cpp VElement.cpp

clean:
	rm -f merge merge_pasos.dat merge_t.dat
