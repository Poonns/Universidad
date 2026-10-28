#include "heuristic.h"

#ifdef ni_HEURISTIC
//esta heuristica fuerza a completar lineas y los tableros con mas posiciones rellenadas 
//t10h resuellto en 547 iteraciones
float h(const Board& next_board)
{
 unsigned x = 0;
 unsigned f;
 enum Token t = e;

    for(unsigned i = 0; i < next_board.GetSize(); i++)
    {
        f = next_board.GetSize()-1;
        for(unsigned j = 0; j < next_board.GetSize(); j++)
        {
            Position p = {i, j};
            if(next_board.Content(p) == t){
                x++;
            }
            else 
            {
                f--;
            }
        }
        if(f <= 0)
            x = x-3; //bonificacion por fila completa
    }
    return (float) x;
}
#endif

#ifdef anyname_HEURISTIC
//esta heuristica fuerza a completar lineas y los tableros con mas posiciones rellenadas 
//t10h resuellto en 547 iteraciones
float h(const Board& next_board)
{
 unsigned x = 0;
 unsigned f;
 enum Token t = e;

    for(unsigned i = 0; i < next_board.GetSize(); i++)
    {
        f = next_board.GetSize()-1;
        for(unsigned j = 0; j < next_board.GetSize(); j++)
        {
            Position p = {i, j};
            if(next_board.Content(p) == t){
                x++;
            }
            else 
            {
                f--;
            }
        }
        if(f <= 0)
            x = x-3; //bonificacion por fila completa
    }
    return (float) x;
}
#endif

// Write here as many heuristics as you want to try, and label each of them with a different name

