all: otello_test

otello_test: otello_test.cpp otello.o minimax.o alphabeta.o heuristic.o
	g++ -Wall otello.o minimax.o alphabeta.o heuristic.o otello_test.cpp -o otello_test

otello.o: otello.cpp otello.h
	g++ -Wall -c otello.cpp

minimax.o: minimax.cpp minimax.h otello.h
	g++ -Wall -c minimax.cpp

alphabeta.o: alphabeta.cpp alphabeta.h otello.h
	g++ -Wall -c alphabeta.cpp

heuristic.o: heuristic.cpp heuristic.h otello.h
	g++ -Wall -c heuristic.cpp

clean:
	-rm -f *.o
	-rm -f otello_test

