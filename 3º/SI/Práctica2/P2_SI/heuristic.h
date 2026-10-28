#ifndef _HEURISTIC_H
#define _HEURISTIC_H

#include "otello.h"

// The heuristic might need this, but it is declared (and initialized) in otello_test.cpp
extern unsigned chosen_depth;
unsigned GetCalls();

// The heuristic function gets the player that is about to play and the current state of the board.
// It must return a measure of how good (or bad) is this state for him/her.
// The convention is that the most favorable the state is, the higher the value of h.
float h(Player pl,Board b);

#endif

