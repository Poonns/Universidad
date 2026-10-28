/**
 * @file merge.cpp
 * @brief Implementación del algoritmo Mergesort y análisis de su coste.
 * * Este programa implementa el algoritmo de ordenación Mergesort utilizando
 * la clase VElement para contar operaciones (asignaciones y comparaciones).
 * Además, incluye funciones para medir el tiempo de ejecución real (CPU time)
 * y generar archivos de datos (.dat) para su posterior análisis gráfico.
 * * @author [Tu Nombre]
 * @date 2025-2026
 */

#include <iostream>
#include <vector>
#include <cstdlib> 
#include <ctime>   
#include <fstream> 
#include "VElement.h"

using namespace std;

//============== Prototipos funciones =================

/**
 * @brief Mezcla dos subvectores ordenados en un vector auxiliar.
 * * Toma dos partes ordenadas del vector v (desde izq hasta m, y desde m+1 hasta der)
 * y las fusiona ordenadamente en el vector aux. Luego copia el resultado de vuelta a v.
 * * @param v Vector principal de elementos VElement.
 * @param aux Vector auxiliar pasado por referencia para evitar copias innecesarias.
 * @param izq Índice izquierdo del subvector actual.
 * @param der Índice derecho del subvector actual.
 * @param m Índice medio que divide los dos subvectores.
 */
void Mezclar(vector<VElement> &v, vector<VElement> &aux, int izq, int der, int m);

/**
 * @brief Función recursiva que divide el vector y ordena las partes.
 * * Implementa la estrategia "Divide y Vencerás":
 * 1. Divide el vector en dos mitades.
 * 2. Se llama a sí misma para ordenar cada mitad.
 * 3. Llama a Mezclar para combinar las mitades ordenadas.
 * * @param v Vector principal a ordenar.
 * @param aux Vector auxiliar para la mezcla.
 * @param ini Índice inicial del rango a ordenar.
 * @param fin Índice final del rango a ordenar.
 */
void MezclarOrd(vector<VElement> &v, vector<VElement> &aux, int ini, int fin);

/**
 * @brief Función genérica para calcular costes y volcarlos a un flujo.
 * * Itera sobre diferentes tamaños de problema (n) y ejecuta una función de medición de coste f(n).
 * Los resultados se escriben en el flujo de salida proporcionado.
 * * @param ini Tamaño inicial del problema.
 * @param fin Tamaño final del problema.
 * @param inc Incremento del tamaño en cada iteración.
 * @param f Puntero a la función de coste que recibe un unsigned (talla) y devuelve un float (coste).
 * @param file Flujo de salida (fichero u otro ostream) donde escribir los pares (n, coste).
 */
void calculoCoste(unsigned ini, unsigned fin, int inc, float (*f)(unsigned), ostream &file);

/**
 * @brief Función Wrapper (envoltorio) para el Mergesort.
 * * Oculta la complejidad de la recursión y la gestión del vector auxiliar al usuario final.
 * Crea el vector auxiliar una única vez antes de iniciar la recursión.
 * * @param v Vector de VElement a ordenar.
 */
void mergesort(vector<VElement> &v);

/**
 * @brief Calcula el coste promedio en "pasos" (operaciones elementales).
 * * Realiza varias ejecuciones para obtener un promedio estadístico.
 * Los "pasos" son la suma de comparaciones y asignaciones registradas por VElement.
 * * @param n Tamaño del vector a ordenar.
 * @return float Promedio de (Asignaciones + Comparaciones).
 */
float costePasosPromedio(unsigned n);

/**
 * @brief Calcula el coste promedio en tiempo real de CPU.
 * * Utiliza clock_gettime para medir el tiempo de proceso con precisión de nanosegundos.
 * Realiza copias del vector para ordenar siempre una secuencia aleatoria idéntica en cada repetición.
 * * @param n Tamaño del vector a ordenar.
 * @return float Promedio del tiempo de ejecución en segundos.
 */
float costeTiempoPromedio(unsigned n);


//============== Función main ==================
/**
 * @brief Punto de entrada principal del programa.
 * * 1. Inicializa la semilla aleatoria.
 * 2. Genera el fichero 'merge_pasos.dat' con el coste en operaciones elementales.
 * 3. Genera el fichero 'merge_t.dat' con el coste en tiempo real.
 * * @return int Código de salida (0 si todo fue correcto).
 */
int main() {
    srand(time(NULL));

    ofstream fichero("merge_pasos.dat");
    if(fichero.is_open()) {
        
        calculoCoste(50, 5000, 50, costePasosPromedio, fichero);
        
        fichero.close();
        cout << "Datos generados correctamente." << endl;
    } else {
        cerr << "Error al crear el fichero." << endl;
    }
    
    ofstream f_tiempo("merge_t.dat");
    
    if(f_tiempo.is_open()) {
        cout << "Generando datos de tiempo (Ejercicio 3)..." << endl;
        cout << "Esto puede tardar un poco porque son muchas repeticiones." << endl;
        
        calculoCoste(50, 5000, 50, costeTiempoPromedio, f_tiempo);
        
        f_tiempo.close();
    } else {
        cerr << "Error al crear el fichero merge_t.dat" << endl;
    }

    cout << "Proceso terminado. Ficheros generados correctamente." << endl;
    
    /* CONCLUSIONES:
     * Tras analizar las gráficas generadas:
     * 1. La curva de pasos (merge_pasos.dat) sigue una tendencia N log N, confirmando
     * el análisis teórico del caso medio.
     * 2. La curva de tiempo real (merge_t.dat), compilando con -O2, también muestra
     * un comportamiento N log N.
     */
     
    return 0;
}

//============ Funcion que mezcla los vectores ==================
void Mezclar(vector<VElement> &v, vector<VElement> &aux, int izq, int der, int m)
{
    int i = izq;
    int j = m + 1;
    int k = izq;
    
    while(i <= m && j <= der)
    {
        if(v[i] <= v[j])
        {
            aux[k] = v[i];
            i++;
        }
        else
        {
            aux[k] = v[j];
            j++;
        }
        
        k++;
    }
    
    while(i <= m)
    {
        aux[k] = v[i];
        i++;
        k++;
    }
    
    while(j <= der)
    {
        aux[k] = v[j];
        j++;
        k++;
    }
    
    for(int t = izq; t <= der; t++)
        v[t] = aux[t];
}

//============ Función mezcla recursiva ======================
void MezclarOrd(vector<VElement> &v, vector<VElement> &aux, int ini, int fin)
{
    if (ini < fin) 
    {
        int m = (ini + fin) / 2;
        MezclarOrd(v, aux, ini, m);
        MezclarOrd(v, aux, m + 1, fin);
        Mezclar(v, aux, ini, fin, m);
    }
}

//=============== Función para Wrapper ========================
void mergesort(vector<VElement> &v)
{
    if(v.empty())
        return;
    
    vector<VElement> aux(v.size());
    
    MezclarOrd(v, aux, 0, v.size() -1);
}


//=============== Calculo coste PASOS =========================
float costePasosPromedio(unsigned n) 
{
    long long sumaPasos = 0;
    int repeticiones = 10; 

    for(int r = 0; r < repeticiones; r++) 
    {
        vector<VElement> v;
        
        v.resize(n);
        for(unsigned k=0; k<n; k++) v[k] = rand();

        VElement::initNAssign();
        VElement::initNComparison();

        mergesort(v);

        sumaPasos += VElement::getNAssign() + VElement::getNComparison();
    }

    return (float)sumaPasos / repeticiones;
}

//================ Pasar los valores a un fichero =====================
void calculoCoste(unsigned ini, unsigned fin, int inc, float (*f)(unsigned), ostream & file)
{
    for(unsigned i = ini; i <= fin; i += inc)
    {
        file << i << " " << f(i) << endl;
    }
}

//=========== Calculo promedio coste TIEMPO ====================
float costeTiempoPromedio(unsigned n)
{
    struct timespec tiempo_ini, tiempo_fin;
    double tiempo_total = 0;
    int repeticiones = 10;
    
    vector<VElement> v(n);
    for(unsigned i = 0; i < n; i++)
        v[i] = rand();
        
    vector<VElement> copia_vec;
    for(unsigned j = 0; j < repeticiones; j++)
    {
        copia_vec = v;
        
        clock_gettime(CLOCK_PROCESS_CPUTIME_ID, &tiempo_ini);
        mergesort(copia_vec);
        clock_gettime(CLOCK_PROCESS_CPUTIME_ID, &tiempo_fin);
        
        double seg = (tiempo_fin.tv_sec - tiempo_ini.tv_sec) + (tiempo_fin.tv_nsec * 1.0e-9 - tiempo_ini.tv_nsec * 1.0e-9);
        tiempo_total += seg;
    }
    
    float promedio = tiempo_total / repeticiones;
    return promedio;
}
