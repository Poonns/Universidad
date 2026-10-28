#include "minimax.h"
#include "heuristic.h"

// This is a global variable, since it might have to be used by heuristic, too.
unsigned chosen_depth=2;

int main(int argc,char *argv[])
{
  float alpha = -FLT_MAX;
  float beta = FLT_MAX;
 if (argc!=1)
 {
  cerr << "This program doesn't take any argument.\n";
  exit(1);
 }

 // The default constructor initializes the board to the state marked by the rules.
 Board b;
 if (DEB)
 {
  cout << "Initial board:\n";
  b.Show(cout);
 }


 // By conventions, Alice starts the game.
 Player pl=Hector;

 // Temporary variables used inside the loop.
 Move mv;
 float score;
 
 // Variables to store the final result, when the game is finished.
 Player winner;
 unsigned winner_score,looser_score;

 // If the game has not fnished (i.e.: at least one of the players can make a move) 
 while (!b.GameOver(winner,winner_score,looser_score))
 {
  // Minimax chooses a move (it will be stored in mv) and returns the score of that move.
  score=AlphaBeta(pl,b,chosen_depth,true,mv, alpha, beta);
 

  if (DEB)
  {
   cout << Name(pl) << " plays. The move is " << "(" << char(mv.second.first) << "," << char(mv.second.second) << ") with a score of ";
   if (score>1E30)
    cout << "+Infinity\n";
   else
    if (score<-1E30)
     cout << "-Infinity\n";
    else 
      cout << score << ".\n";
  }

  // The move is actually executed (the state of the board is internally changed)
  b.MakeMove(mv);

  if (DEB)
  {
   cout << "The board is now\n";
   b.Show(cout);
  }
  
  // Next player is the opponent of the current one.
  pl=Opponent(pl);
 }

 if (DEB)
 {
  cout << "Las llamadas son " << GetCalls() << endl;
  cout << "The winner is " << Name(winner) << " with a score of " << winner_score << " against the " << looser_score << " of " << Name(Opponent(winner)) << ".\n";
  if (winner_score==looser_score)
   cout << "Well, yeah, I know this is a tie, but in such a case ladies win, doesn't do?\n";
 }
}

