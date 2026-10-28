#ifndef _MINIMAX_H
#define _MINIMAX_H

#include "otello.h"
float AlphaBeta(Player pl,Board b,unsigned depth,bool maximizingPlayer,Move &best_move, float alpha, float beta);
#endif

