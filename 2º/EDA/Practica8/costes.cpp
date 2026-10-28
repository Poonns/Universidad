/**
 * @file main.cpp
 * @brief Programa para comparar el rendimiento de algoritmos Voraz vs Backtracking en el TSP restringido.
 *
 * Este programa realiza un estudio experimental generando grafos aleatorios de diferentes tamanyos.
 * Ejecuta ambos algoritmos, mide el numero de operaciones elementales (accesos) y la calidad de la solucion,
 * y vuelca los resultados en archivos .dat para su posterior analisis grafico (por ejemplo con Gnuplot).
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
#include <cstdlib>

using namespace std;

/**
 * @brief Constante para representar infinito.
 */
const int INF = INT_MAX;

// -------------------------------------------------------------------------
// Variables Globales para Metricas
// -------------------------------------------------------------------------

/**
 * @brief Contador de operaciones (accesos a la matriz/comparaciones) para el algoritmo Voraz.
 */
long long accesos_voraz = 0;

/**
 * @brief Contador de operaciones (accesos a la matriz/nodos visitados) para el algoritmo Backtracking.
 */
long long accesos_bt = 0;

/**
 * @brief Almacena el mejor coste encontrado durante la recursion de Backtracking.
 */
int mejor_tiempo_bt = INF;

/**
 * @brief (Opcional) Almacena el mejor camino encontrado por Backtracking.
 * En este codigo se declara pero no se utiliza intensivamente en la salida de datos.
 */
vector<int> mejor_camino_bt;

/**
 * @brief Genera un grafo completo aleatorio (simetrico o asimetrico segun implementacion).
 *
 * Los pesos de las aristas se generan aleatoriamente entre 30 y 240.
 * La diagonal principal se inicializa a 0.
 *
 * @param n Numero de nodos del grafo.
 * @return vector<vector<int>> Matriz de adyacencia (n x n) con los costes.
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
 * @brief Ejecuta la heuristica Voraz (Greedy) para el TSP restringido.
 *
 * Estrategia: Desde el nodo actual, viaja al vecino no visitado con menor coste,
 * respetando la restriccion de no ir al nodo consecutivo (i+1).
 *
 * @param n Numero de nodos.
 * @param G Matriz de adyacencia.
 * @return int Coste total del camino encontrado, o -1 si no se encontro solucion factible.
 */
int ejecutar_voraz(int n, const vector<vector<int>>& G) 
{
    accesos_voraz = 0;
    vector<bool> visitado(n, false);
    int tiempo_total = 0;
    int actual = 0;

    visitado[actual] = true;

    for (int i = 1; i < n; i++) 
    {
        int mejor_vecino = -1;
        int menor_coste = INF;

        for (int j = 0; j < n; j++) 
        {
            accesos_voraz++; // Contabilizamos acceso a la matriz/evaluacion
            
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
            return -1; 

        actual = mejor_vecino;
        visitado[actual] = true;
        tiempo_total += menor_coste;
    }

    accesos_voraz++; 
    
    if (0 == actual + 1) return -1; 

    tiempo_total += G[actual][0];
    return tiempo_total;
}

/**
 * @brief Funcion recursiva auxiliar para el algoritmo de Backtracking.
 *
 * Explora el arbol de soluciones posibles utilizando poda por coste.
 *
 * @param n Numero de nodos.
 * @param G Matriz de adyacencia.
 * @param camino Vector con el camino actual en construccion.
 * @param visitado Vector booleano para control de nodos visitados.
 * @param tiempo_acumulado Coste acumulado hasta el nivel actual.
 * @param nivel Numero de nodos incluidos en el camino actual.
 */
void backtracking_rec(int n, const vector<vector<int>>& G, vector<int>& camino, vector<bool>& visitado, int tiempo_acumulado, int nivel) 
{
    if (nivel == n) 
    {
        int ultimo = camino.back();
        if (0 != ultimo + 1) 
        {
            accesos_bt++;
            int tiempo_final = tiempo_acumulado + G[ultimo][0];
            
            if (tiempo_final < mejor_tiempo_bt) 
            {
                mejor_tiempo_bt = tiempo_final;
            }
        }
        return;
    }

    int actual = camino.back();
    for (int j = 0; j < n; j++) 
    {
        if (!visitado[j] && (j != actual + 1)) 
        {
            accesos_bt++;
            int nuevo_tiempo = tiempo_acumulado + G[actual][j];

            if (nuevo_tiempo < mejor_tiempo_bt) 
            {
                visitado[j] = true;
                camino.push_back(j);
                
                backtracking_rec(n, G, camino, visitado, nuevo_tiempo, nivel + 1);

                camino.pop_back();
                visitado[j] = false;
            }
        }
    }
}

/**
 * @brief Funcion envoltorio (wrapper) para preparar y ejecutar Backtracking.
 *
 * Inicializa las variables globales y estructuras necesarias antes de llamar a la recursion.
 *
 * @param n Numero de nodos.
 * @param G Matriz de adyacencia.
 * @return int El coste del mejor camino encontrado (o INF si no existe).
 */
int ejecutar_bt(int n, const vector<vector<int>>& G) 
{
    accesos_bt = 0;
    mejor_tiempo_bt = INF;
    vector<bool> visitado(n, false);
    vector<int> camino;
    
    visitado[0] = true;
    camino.push_back(0);

    backtracking_rec(n, G, camino, visitado, 0, 1);
    
    return mejor_tiempo_bt;
}

/**
 * @brief Funcion principal. Gestiona la experimentacion y generacion de archivos.
 *
 * 1. Crea archivos .dat para guardar los resultados.
 * 2. Itera para diferentes tamanyos de grafo (n_min a n_max).
 * 3. Para cada tamanyo, ejecuta 'repeticiones' veces generando grafos aleatorios.
 * 4. Calcula promedios de accesos y calidad de la solucion.
 * 5. Escribe los resultados en los archivos.
 *
 * @return int 0 si la ejecucion es correcta, 1 si hay error de ficheros.
 */
int main() 
{
    srand(time(NULL));
    ofstream f_voraz_coste("voraz.dat");
    ofstream f_bt_coste("bt.dat");
    ofstream f_voraz_sol("voraz_sol.dat");
    ofstream f_bt_sol("bt_sol.dat");

    if (!f_voraz_coste || !f_bt_coste || !f_voraz_sol || !f_bt_sol) 
    {
        cerr << "Error creando archivos .dat" << endl;
        return 1;
    }

    int repeticiones = 10;
    int n_min = 5;
    int n_max = 10;

    cout << "Generando datos para N [" << n_min << ", " << n_max << "] con " << repeticiones << " repeticiones..." << endl;

    for (int n = n_min; n <= n_max; n++) 
    {
        long long suma_accesos_v = 0;
        long long suma_accesos_bt = 0;
        long long suma_sol_v = 0;
        long long suma_sol_bt = 0;
        
        int exitos_v = 0;
        int exitos_bt = 0;

        for (int r = 0; r < repeticiones; r++) 
        {
            vector<vector<int>> G = generarGrafo(n);

            int coste_v = ejecutar_voraz(n, G);
            suma_accesos_v += accesos_voraz;
            if (coste_v != -1) 
            {
                suma_sol_v += coste_v;
                exitos_v++;
            } 
            else 
            {
                suma_sol_v += (240 * n); 
            }

            int coste_bt = ejecutar_bt(n, G);
            suma_accesos_bt += accesos_bt;
            if (coste_bt != INF) 
            {
                suma_sol_bt += coste_bt;
                exitos_bt++;
            }
        }

        double media_acc_v = (double)suma_accesos_v / repeticiones;
        double media_acc_bt = (double)suma_accesos_bt / repeticiones;
        
        double media_sol_v = (double)suma_sol_v / repeticiones;
        double media_sol_bt = (double)suma_sol_bt / repeticiones;

        f_voraz_coste << n << " " << media_acc_v << endl;
        f_bt_coste << n << " " << media_acc_bt << endl;
        
        f_voraz_sol << n << " " << media_sol_v << endl;
        f_bt_sol << n << " " << media_sol_bt << endl;

        cout << "N=" << n << " completado." << endl;
    }

    f_voraz_coste.close();
    f_bt_coste.close();
    f_voraz_sol.close();
    f_bt_sol.close();

    cout << "Archivos .dat generados correctamente." << endl;

    return 0;
}