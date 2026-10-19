#include "takuzu.h"
#include "heuristic.h"

// This is a global variable with the tables of g and f functions. Don't erase it
Scores S;


//Función que dice si es un estado final
bool esFinal(Board b)
{

    enum Token t = e;

    for(unsigned i = 0; i < b.GetSize(); i++)
    {
        
        for(unsigned j = 0; j < b.GetSize(); j++)
        {
            Position p = {i, j};
            if(b.Content(p) == t)
                return false;
        }
    }
    return true;
}
// The effort to move from any state of the game to any of its neighbours is considered to be the same,
// independently of the chosen neighbour. This is just a choice; more sophisticated distances can be
// programmed, but we prefer to subsume them inside the heuristics so don't change this function
float dist(Board b1,Board b2)
{
 return(1.0);
}

// This function must be done by the alumni
// It should return true and set the variable 'solved' to the full takuzu final board,
// or return false and set the variable 'solved' to the last board tried.
bool AStar(Board start,Board &solved)
{
 Board::lb open;
 Board::lb closed;
 Board::lnei nei;
 Board current(start.GetSize());


 open.push_back(start);

 Move mov_nulo;
 mov_nulo.sq.col = 0;
 mov_nulo.sq.row = 0;
 mov_nulo.value = e;
 
 S.SetgValue(start, 0.0);
 S.SetfValue(start, 0.0 + h(start, mov_nulo, start));

 while(! open.empty())
 {
    Board::lb::iterator mejor_it = open.begin();

    for (Board::lb::iterator it = open.begin(); it != open.end(); it++) {
        if (S.GetfValue(*it) < S.GetfValue(*mejor_it)) {
            mejor_it = it; 
        }
    }

    current = *mejor_it;  
    open.erase(mejor_it);

    if(esFinal(current))
    {
        solved = current;
        return true;
    }  
    
    closed.push_back(current); 
    
    current.GetNeighbours(nei);
    for (auto &par : nei) 
    {

        Move m = par.first;
        Board vecino_board = par.second;

        if (!vecino_board.Member(closed)) 
        {
            float g_provisional = S.GetgValue(current) + dist(current, vecino_board);
            if (g_provisional < S.GetgValue(vecino_board)) 
            {
                S.SetgValue(vecino_board, g_provisional);
                S.SetfValue(vecino_board, g_provisional + h(current, m, vecino_board));
                if (!vecino_board.Member(open)) 
                {
                    open.push_back(vecino_board); 
                }
            }
        }
    }
 }
 solved = current;
 return false;
}

// Don't change main
int main(int argc,char *argv[])
{
 if (argc!=2)
 {
  cerr << "Usage:\n       " << argv[0] << "board_file\n";
  exit(1);
 }

 string fname=string(argv[1]);

 // A board is created and filled with the content of the file passed as argument
 // By using the constructor with two arguments we say that we want to print in the console (cout)
 // the initial board, as loaded from the file, and the board after the compulsory movements have been done
 Board my_board(fname,cout);
 
 // Now, using the constructor with one argument (the number of squares) we create an empty board to contain the final solution
 Board solution(my_board.GetSize());
 
 if (AStar(my_board,solution))
  cout << "Takuzu solved!. The final board is:\n";
 else
  cout << "Takuzu NOT solved!. The current board is\n";
  
 // We print in the screen the final board, either if a solution has been found or if not
 solution.Show(cout);
  
 return 0;
}
