/**
 * @file Grafo.h
 * @brief Declaracion de la clase Grafo para el manejo de grafos ponderados mediante matriz de adyacencia.
 */
#ifndef GRAFO_H
#define GRAFO_H

#include <vector>
#include <iostream>
#include <stdexcept>

using namespace std;

class Grafo 
{
public:
    Grafo(unsigned n);
    float& operator()(unsigned i, unsigned j);
    float operator()(unsigned i, unsigned j) const;
    void creaGrafoND(float pmin, float pmax);
    unsigned getAccesos() const;
    void resetAccesos() const;
    unsigned talla() const;
    void mostrar() const;

private:
    unsigned n; ///< Numero de nodos del grafo
    vector<vector<float>> matriz; ///< Matriz de adyacencia que almacena los pesos de las aristas
    mutable unsigned num_accesos; ///< Contador de accesos a la matriz 
};

#endif