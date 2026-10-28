/**
 * @file main_voraz.cpp
 * @brief Implementacion del algoritmo Voraz (Greedy) para el TSP con restricciones.
 *
 * Este programa resuelve el problema del viajante de comercio con la restriccion de
 * no viajar de un nodo i a un nodo i+1. Ademas, realiza un analisis empirico
 * del coste computacional contando los accesos a la matriz de adyacencia
 * y comparandolos con la complejidad teorica cuadratica.
 *
 * @author Hector Pons
 * @author Meryem Bagho
 */

#include <iostream>
#include <vector>
#include <climits>
#include <cassert>
#include <fstream>
#include <ctime>

using namespace std;

/**
 * @brief Constante para representar un costo infinito.
 */
const int INF = INT_MAX;

/**
 * @brief Variable global para contabilizar accesos a la matriz de adyacencia.
 * Se utiliza para medir la eficiencia empirica del algoritmo.
 */
long long accesos = 0;

/**
 * @brief Verifica la validez de una solucion propuesta (camino).
 *
 * Comprueba que la solucion sea un ciclo Hamiltoniano valido y respete
 * la restriccion especifica del problema.
 *
 * @param n Numero total de nodos.
 * @param sol Vector con la secuencia de nodos.
 * @return true Si la solucion es correcta.
 * @return false Si la solucion es invalida.
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

/**
 * @brief Estructura para almacenar el resultado de la ejecucion voraz.
 */
struct ResultadoVoraz 
{
    vector<int> camino; ///< Secuencia de nodos visitados.
    int coste_total;    ///< Coste acumulado del recorrido.
};

/**
 * @brief Algoritmo Voraz (Vecino mas cercano) con restricciones.
 *
 * Parte del nodo 0 y en cada paso selecciona el vecino no visitado con menor coste,
 * siempre que cumpla la restriccion de no ser el consecutivo numerico (actual + 1).
 *
 * @param n Numero de nodos del grafo.
 * @param G Matriz de adyacencia con los pesos de las aristas.
 * @return ResultadoVoraz Estructura con el camino y el coste. Si no encuentra solucion, devuelve coste maximo.
 */
ResultadoVoraz algoritmo_voraz(int n, const vector<vector<int>>& G) 
{
    assert(n > 0); 
    
    vector<int> camino;
    vector<bool> visitado(n, false);
    int coste_acumulado = 0;
    int actual = 0;  

    camino.push_back(actual);
    visitado[actual] = true;

    for (int i = 1; i < n; i++) 
    {
        int mejor_vecino = -1;
        int menor_coste = INF;

        for (int j = 0; j < n; j++) {
            accesos++; 
            
            if (!visitado[j] && (j != actual + 1)) 
            {
                if (G[actual][j] < menor_coste) 
                {
                    menor_coste = G[actual][j];
                    mejor_vecino = j;
                }
            }
        }

        if (mejor_vecino == -1) 
        {
            return {vector<int>(), 240 * n}; 
        }

        actual = mejor_vecino;
        camino.push_back(actual);
        visitado[actual] = true;
        coste_acumulado += menor_coste;
    }

    // Regreso al origen
    accesos++;
    
    if (0 == actual + 1) 
    {
        return {vector<int>(), 240 * n};
    }
    
    camino.push_back(0);
    coste_acumulado += G[actual][0];

    return {camino, coste_acumulado};
}

/**
 * @brief Genera un grafo completo con pesos aleatorios.
 *
 * @param n Numero de nodos.
 * @return vector<vector<int>> Matriz de adyacencia con pesos entre 30 y 240.
 */
vector<vector<int>> generarGrafo(int n) 
{
    vector<vector<int>> G(n, vector<int>(n));
    for (int i = 0; i < n; i++) 
    {
        for (int j = 0; j < n; j++) 
        {
            if (i == j) G[i][j] = 0;
            else G[i][j] = 30 + rand() % (240 - 30 + 1);
        }
    }
    return G;
}

/**
 * @brief Funcion principal del programa.
 *
 * Realiza dos tareas:
 * 1. Ejecuta una prueba unitaria con n=5 e imprime el resultado y su validacion.
 * 2. Ejecuta un bucle de pruebas de n=5 a n=100 para generar un archivo 'voraz.dat'
 * con el numero medio de accesos realizados vs el numero teorico (n^2).
 *
 * @return int Codigo de salida.
 */
int main() 
{
    srand(time(NULL));
    
    int n_fijo = 5;
    vector<vector<int>> grafo = generarGrafo(n_fijo);

    accesos = 0;
    ResultadoVoraz res = algoritmo_voraz(n_fijo, grafo);

    if (res.camino.empty()) 
    {
        cout << "No se encontro solucion. Tiempo: " << res.coste_total << endl;
    } 
    else 
    {
        cout << "Solucion Voraz (n=5): ";
        for (int nodo : res.camino) 
            cout << nodo << " ";
        cout << "\nTiempo total: " << res.coste_total << " segundos." << endl;

        if (Solucion_OK(n_fijo, res.camino)) 
            cout << "La solucion es CORRECTA." << endl;
        else 
            cout << "La solucion es INVALIDA." << endl;
    }
    ofstream f_dat("voraz.dat");

    for (int n = 5; n <= 100; n += 5) 
    {
        long long total_accesos = 0;
        int reps = 10; 
        for (int r = 0; r < reps; r++) 
        {
            vector<vector<int>> G = generarGrafo(n);
            accesos = 0;
            algoritmo_voraz(n, G);
            total_accesos += accesos;
        }
        
        f_dat << n << " " << (double)total_accesos / reps << " " << n * n << endl;
    }
    f_dat.close();

    return 0;
}