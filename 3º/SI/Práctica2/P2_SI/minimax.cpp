#include "minimax.h"
#include "heuristic.h"

/*
float Minimax(Player pl,Board b,unsigned depth,bool maximizingPlayer,Move &best_move)

Parameters:
Player pl:             the player who is about to put a token

Board b:               the current state of the board

unsigned depth:        the current depth. In the first call it should be passed as the chosen
                       depth. In further recursive calls it is diminished so when it arrives to 0
                       it plays the role of a signal to warn us of not going on searching.

bool maximizingPlayer: boolean variable to indicate if this is the player who wants to
                       maximize or to minimize. When a player calls Minimax to choose his/her
                       move, it should be passed as true (he/she is trying to get the maximum
                       value of the function) but when Minimax is recursively called, it should
                       be passed as false if we have come here with true (since a minimization
                       step follows a maximization one) and conversely. See the actual code.

Move &best_move:       Variable to return the best move we have found at this level. The best move
                       for deeper levels is found, too, but not returned, since we are only interested
                       in what to do from here, not in the prospective decisions we think our opponent
                       will take (even such decisions have lead us to arrive to our final decision).
                       Again, see actual code. 

Returns:
			The optimal value of the evaluation function
*/

float AlphaBeta(Player pl,Board b,unsigned depth,bool maximizingPlayer,Move &best_move, float alpha, float beta)
{
     
 Sqlist sons=b.ValidMoves(pl);

 if ( depth==0 || sons.empty() )
  return h(pl,b);

 float best_value;

 if (maximizingPlayer)
 {

  best_value=-FLT_MAX;

  float v;
  Move mv;

  for (Sqlist::iterator it=sons.begin();it!=sons.end();it++)
  {

   mv=make_pair(pl,(*it));

   bool legal=false;
   Board child=b.TryMove(mv,legal);
   if (!legal)
   {

    cerr << "Error from Minimax: trying illegal movement.\n";
    exit(1);
   }

   Move dummy_move;
   v=AlphaBeta(Opponent(pl),child,depth-1,false,dummy_move,alpha,beta);

   if (v>best_value)
   {
    best_value=v;
    best_move=mv;
   }
   if (v > alpha)
   {
    alpha = v;
   }
   if (beta <= alpha)
   {
    return best_value;
   }
  }
  return best_value;
 }

 else
 {

  best_value=FLT_MAX;
  float v;
  Move mv;
  for (Sqlist::iterator it=sons.begin();it!=sons.end();it++)
  {
   mv=make_pair(pl,(*it));
   bool legal=false;
   Board child=b.TryMove(mv,legal);
   if (!legal)
   {
    cerr << "Error from Minimax: trying illegal movement.\n";
    exit(1);
   }

   Move dummy_move;
   v=AlphaBeta(Opponent(pl),child,depth-1,true,dummy_move,alpha,beta);

   if (v<best_value)
   {
    best_value=v;
    best_move=mv;
   }
   if (v < beta)
   {
    beta = v;
   }
   if (beta <= alpha)
   {
    return best_value;
   }
  }
  return best_value; 
 }
}  


