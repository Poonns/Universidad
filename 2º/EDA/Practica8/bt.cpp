/**
 * @file main.cpp
 * @brief Solucion al Problema del Viajante de Comercio (TSP) con restricciones usando Backtracking.
 *
 * Este programa resuelve una variante del TSP donde se busca el ciclo Hamiltoniano de menor costo
 * sujeto a una restriccion especifica: No se puede viajar de un nodo i a un nodo i+1.
 * Utiliza un algoritmo de vuelta atras (backtracking) con poda por costo.
 *
 * @author Hector Pons
 * @author Meryem Bagho
 */

#include <iostream>
#include <vector>
#include <climits>
#include <iomanip> 
#include <cassert>

using namespace std;

/**
 * @brief Valor constante para representar un costo infinito inicial.
 */
const int INF = INT_MAX;

/**
 * @brief Almacena la secuencia de nodos que conforman la mejor solucion encontrada hasta el momento.
 */
vector<int> mejor_camino;

/**
 * @brief Almacena el costo total (tiempo) de la mejor solucion encontrada.
 * Se inicializa en INF para permitir la minimizacion.
 */
int mejor_tiempo = INF;

/**
 * @brief Contador de accesos a nodos del arbol de exploracion (metricas de rendimiento).
 */
long long accesos_bt = 0; 

/**
 * @brief Valida si una solucion propuesta cumple con todas las reglas del problema.
 *
 * Verifica lo siguiente:
 * 1. El tamanyo del camino debe ser n + 1 (ciclo cerrado).
 * 2. Debe empezar y terminar en el nodo 0.
 * 3. Todos los nodos deben ser visitados exactamente una vez (excepto el origen/destino).
 * 4. Se debe respetar la restriccion: no existen arcos directos de i -> i+1.
 *
 * @param n Numero total de nodos en el grafo.
 * @param sol Vector que contiene la secuencia de nodos de la solucion a validar.
 * @return true Si la solucion es valida y correcta.
 * @return false Si la solucion viola alguna restriccion o estructura.
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
 * @brief Algoritmo de Backtracking para encontrar el camino de costo minimo.
 *
 * Explora recursivamente las posibles rutas. Incluye dos mecanismos de optimizacion/control:
 * 1. **Poda:** Si el costo actual supera al mejor costo encontrado (mejor_tiempo), se abandona la rama.
 * 2. **Restriccion:** No explora arcos donde el destino sea origen + 1.
 *
 * @param n Numero total de nodos del grafo.
 * @param G Matriz de adyacencia que representa los costos entre nodos.
 * @param actual_camino Vector que mantiene el camino construido en la recursion actual.
 * @param visitado Vector de control para marcar nodos ya incluidos en el camino actual.
 * @param tiempo_acumulado Suma de los costos de los arcos recorridos hasta el momento.
 * @param nivel Profundidad actual en el arbol de recursion (numero de nodos visitados).
 */
void backtracking(int n, const vector<vector<int>>& G, vector<int>& actual_camino, vector<bool>& visitado, int tiempo_acumulado, int nivel) 
{
    if (nivel == n) 
    {
        int ultimo = actual_camino.back();
        
        if (0 != ultimo + 1) 
        {
            accesos_bt++; 
            int tiempo_final = tiempo_acumulado + G[ultimo][0];
            
            if (tiempo_final < mejor_tiempo) 
            {
                mejor_tiempo = tiempo_final;
                mejor_camino = actual_camino;
                mejor_camino.push_back(0); 
            }
        }
        return;
    }

    int actual_nodo = actual_camino.back();

    for (int j = 0; j < n; j++) 
    {
        if (!visitado[j] && (j != actual_nodo + 1)) 
        {
            accesos_bt++; 
            int coste_arco = G[actual_nodo][j];
            
            if (tiempo_acumulado + coste_arco < mejor_tiempo) 
            {
                visitado[j] = true;
                actual_camino.push_back(j);

                backtracking(n, G, actual_camino, visitado, tiempo_acumulado + coste_arco, nivel + 1);

                actual_camino.pop_back();
                visitado[j] = false;
            }
        }
    }
}

/**
 * @brief Imprime la matriz de adyacencia del grafo por consola de forma formateada.
 *
 * @param n Dimension de la matriz (n x n).
 * @param G Referencia constante a la matriz de adyacencia.
 */
void mostrarGrafo(int n, const vector<vector<int>>& G) 
{
    cout << "GRAFO (Matriz de adyacencia)" << endl;
    for (int i = 0; i < n; i++) 
    {
        for (int j = 0; j < n; j++) 
            cout << setw(4) << G[i][j] << " ";
        cout << endl;
    }
}

/**
 * @brief Funcion principal del programa.
 *
 * Define el grafo de ejemplo, inicializa las estructuras de datos necesarias
 * e invoca al algoritmo de backtracking. Finalmente, imprime los resultados y valida la solucion.
 *
 * @return int Codigo de salida del programa (0 si es exitoso).
 */
int main() 
{
    int n = 5;
    
    vector<vector<int>> G = 
    {
        {0,   100, 150, 200, 50},
        {100, 0,   30,  80,  120},
        {150, 30,  0,   40,  90},
        {200, 80,  40,  0,   60},
        {50,  120, 90,  60,  0}
    };

    mostrarGrafo(n, G);

    vector<int> camino_actual;
    camino_actual.push_back(0);
    
    vector<bool> visitado(n, false);
    visitado[0] = true; 

    // Reinicializar globales
    mejor_tiempo = INF;
    mejor_camino.clear();
    accesos_bt = 0;

    backtracking(n, G, camino_actual, visitado, 0, 1);

    if (mejor_tiempo == INF)
        cout << "No se encontro solucion factible." << endl;
    
    else 
    {
        cout << "\nSolucion Backtracking:" << endl;
        cout << "Camino: ";
        for (size_t i = 0; i < mejor_camino.size(); i++) 
            cout << mejor_camino[i] << (i < mejor_camino.size() - 1 ? " -> " : "");
        
        cout << endl;
        cout << "Tiempo total: " << mejor_tiempo << " segundos." << endl;
        cout << "Accesos al grafo: " << accesos_bt << endl;

        if (Solucion_OK(n, mejor_camino)) 
            cout << "VALIDACION: La solucion es CORRECTA." << endl;
        else
            cout << "VALIDACION: La solucion es INCORRECTA." << endl;
        
    }

    return 0;
}