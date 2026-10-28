/**
 * @file Grafo.cpp
 * @brief Implementacion de los metodos de la clase Grafo.
 */
#include "Grafo.h"
#include <iomanip>
#include <cstdlib>
#include <ctime>

/**
 * @brief Constructor de la clase Grafo.
 * * Inicializa la matriz de n x n con ceros y el contador de accesos a 0.
 * * @param n Numero de nodos del grafo.
 */
Grafo::Grafo(unsigned n) : n(n), num_accesos(0)
{
    matriz.resize(n);
    for (unsigned i = 0; i < n; ++i)
        matriz[i].resize(n, 0.0f);
}

/**
 * @brief Operador de acceso a la matriz (lectura/escritura).
 * Incrementa el contador de accesos y devuelve una referencia al peso de la arista.
 * @param i Nodo origen.
 * @param j Nodo destino.
 * @return float& Referencia al peso de la arista.
 * @throw std::out_of_range Si i o j son mayores o iguales a n.
 */
float& Grafo::operator()(unsigned i, unsigned j)
{
    num_accesos++;
    if (i >= n || j >= n)
        throw out_of_range("Indice de nodo fuera de rango");
    return matriz[i][j];
}

/**
 * @brief Operador de acceso a la matriz (solo lectura).
 * Incrementa el contador de accesos y devuelve el valor del peso de la arista.
 * @param i Nodo origen.
 * @param j Nodo destino.
 * @return float Valor del peso de la arista.
 * @throw std::out_of_range Si i o j son mayores o iguales a n.
 */
float Grafo::operator()(unsigned i, unsigned j) const
{
    num_accesos++;
    if (i >= n || j >= n)
        throw out_of_range("Indice de nodo fuera de rango");
    return matriz[i][j];
}

/**
 * @brief Rellena el grafo con pesos aleatorios para crear un grafo No Dirigido.
 * Recorre la triangular superior de la matriz, asigna un peso aleatorio
 * y lo refleja en la posicion simetrica (matriz[j][i]) para garantizar que sea no dirigido.
 * @param pmin Valor mínimo del peso aleatorio.
 * @param pmax Valor máximo del peso aleatorio.
 */
void Grafo::creaGrafoND(float pmin, float pmax)
{
    float r, peso;
    for (unsigned i = 0; i < n; ++i)
    {
        matriz[i][i] = 0.0f;
        for (unsigned j = i + 1; j < n; ++j)
        {
            r = rand() / float(RAND_MAX);
            peso = pmin + r * (pmax - pmin);            
            matriz[i][j] = peso;
            matriz[j][i] = peso;
        }
    }
}

/**
 * @brief Devuelve el numero acumulado de accesos a la matriz.
 * @return unsigned Contador de accesos.
 */
unsigned Grafo::getAccesos() const 
{
    return num_accesos;
}

/**
 * @brief Reinicia el contador de accesos a cero.
 */
void Grafo::resetAccesos() const 
{
    num_accesos = 0;
}

/**
 * @brief Devuelve la dimension del grafo (numero de nodos).
 * @return unsigned Talla del grafo.
 */
unsigned Grafo::talla() const 
{
    return n;
}

/**
 * @brief Muestra la matriz de adyacencia por la salida estandar.
 * Formatea la salida con 1a decimal de precision y separadores de tabulacion.
 */
void Grafo::mostrar() const
{
    cout << fixed << setprecision(1);
    for (unsigned i = 0; i < n; ++i)
    {
        for (unsigned j = 0; j < n; ++j)
            cout << matriz[i][j] << "\t"; 
        cout << endl; 
    }
}