/**
 * @file Prim.h
 * @brief Implementacion del algoritmo de Prim para encontrar el arbol de expansion minima.
 */
#ifndef PRIM_H
#define PRIM_H

#include "Grafo.h"
#include <vector>
#include <iostream>

const float INF = 1000000000.0f;
using namespace std;

/**
 * @struct ResultadoPrim
 * @brief Estructura que almacena el resultado de la ejecucion del algoritmo de Prim.
 * Contiene el coste total del arbol y las aristas que lo componen.
 */
struct ResultadoPrim
{
    float coste_total;
    vector<pair<unsigned, unsigned>> arcos;
};

/**
 * @brief Ejecuta el algoritmo de Prim para encontrar el arbol de Expansion Minima.
 * El algoritmo comienza desde el nodo 0 y busca iterativamente el nodo no visitado más cercano al conjunto de nodos ya visitados, actualizando los costes mínimos
 * de sus vecinos.
 * @param g Referencia constante al grafo sobre el cual se aplicara el algoritmo.
 * @return ResultadoPrim Estructura con el coste total acumulado y el vector de aristas del arbol resultante.
 */
ResultadoPrim prim(const Grafo& g)
{
    ResultadoPrim res;
    int u;
    float min_val, peso;
    unsigned n = g.talla();

    if (n == 0) 
        return {0.0f, {}};
    vector<bool> visitado(n, false);
    vector<float> costeMinimo(n, INF);
    vector<int> padre(n, -1);

    costeMinimo[0] = 0.0f;

    res.coste_total = 0.0f;

    for (unsigned i = 0; i < n; ++i)
    {
        u = -1;
        min_val  = INF;

        for (unsigned v = 0; v < n; ++v)
        {
            if (!visitado[v] && costeMinimo[v] < min_val)
            {
                min_val = costeMinimo[v];
                u = v;
            }
        }
        if (u == -1)
            break;
        visitado[u] = true;

        if (u != 0)
        {
            res.arcos.push_back({(unsigned)padre[u], (unsigned)u});
            res.coste_total += g((unsigned)padre[u], (unsigned)u);
        }
        for (unsigned v = 0; v < n; ++v)
        {
            if (!visitado[v])
            {
                peso = g((unsigned)u, v);
                if (peso < costeMinimo[v])
                {
                    costeMinimo[v] = peso;
                    padre[v] = u;
                }
            }
        }
    }
    return res;
}
#endif