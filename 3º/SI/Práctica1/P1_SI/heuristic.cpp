#include "heuristic.h"

#ifdef ni_HEURISTIC
// The non-informative heuristic
float h(Board current_board,Move m,Board next_board)
{
 return 0.0;// Program here the non-informative heuristica
}
#endif

#ifdef anyname_HEURISTIC
float h(Board current_board,Move m,Board next_board)
{
 return 0.0;// Give a name to this heuritic and program it
}
#endif

// Write here as many heuristics as you want to try, and label each of them with a different name

