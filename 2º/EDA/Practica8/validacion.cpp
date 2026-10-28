/**
 * @file validacion.cpp
 * @brief Modulo de validacion para soluciones del problema del viajante.
 * @author Hector Pons
 * @author Meryem Bagho
 */

#include <vector>
#include <iostream>

using namespace std;

/**
 * @brief Verifica la validez de una solucion propuesta (camino).
 *
 * Comprueba que la solucion cumpla con las reglas estructurales de un ciclo Hamiltoniano
 * y con la restriccion especifica del problema (prohibido paso de i a i+1).
 *
 * Las comprobaciones son:
 * 1. Longitud del vector debe ser n + 1 (para cerrar el ciclo).
 * 2. El primer y ultimo elemento deben ser el nodo 0 (origen).
 * 3. Los identificadores de nodo deben estar en el rango [0, n-1].
 * 4. No deben existir nodos repetidos en los primeros n elementos (visitado exacto).
 * 5. Restriccion: El nodo siguiente no puede ser (actual + 1).
 *
 * @param n Numero total de nodos del grafo.
 * @param sol Vector de enteros que representa la secuencia de nodos visitados.
 * @return true Si la solucion es valida y factible.
 * @return false Si viola alguna restriccion o estructura basica.
 */
bool Solucion_OK(int n, const vector<int>& sol) 
{
    if (sol.size() != (size_t)n + 1) 
        return false; 

    if (sol.front() != 0 || sol.back() != 0) 
        return false; 

    vector<bool> visitado(n, false);
    for (int i = 0; i < n; i++) 
    {
        int actual = sol[i];
        int siguiente = sol[i + 1];

        if (actual < 0 || actual >= n) 
            return false;
        
        if (visitado[actual]) 
            return false;
        
        visitado[actual] = true;

        if (siguiente == actual + 1) 
            return false;
    }
    
    return true;
}