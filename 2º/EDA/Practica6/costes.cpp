/**
 * @file costes.cpp
 * @brief Programa para analizar el rendimiento (complejidad temporal) de diferentes estructuras de datos.
 * @details Compara los tiempos de ejecución de búsqueda en:
 * 1. ABB (Árbol Binario de Búsqueda).
 * 2. std::map (Implementación de árbol rojo-negro de la STL).
 * 3. Tabla Hash (Implementación propia).
 * @date Curso 2025/2026
 * @author Meryem Bagho
 * @author Hector Pons
 */

#include <iostream>
#include <fstream>
#include <vector>
#include <ctime>
#include <map>
#include "tabla.h"
#include "alumno.h"
#include "abb.h"

using namespace std;

// PROTOTIPOS
void calculoCoste(unsigned ini, unsigned fin, int inc, void (*f)(unsigned), ostream& file);
void costeABB(unsigned);
void costeMap(unsigned);
void costeTabla(unsigned);

int main (void)
{
    srand(time(NULL)); 
    
    ofstream fichero_abb("abb.dat");
    ofstream fichero_map("map.dat");
    ofstream fichero_tabla("tabla.dat");
    
    // --- INICIO DE LAS PRUEBAS ---

    cout << "Calculando costes ABB..." << endl;
    // Pasamos la función 'costeABB' como parámetro (puntero a función)
    calculoCoste(500, 5000, 500, costeABB, fichero_abb);
    
    cout << "Calculando costes Map..." << endl;
    calculoCoste(500, 5000, 500, costeMap, fichero_map);
    
    cout << "Calculando costes Tabla Hash..." << endl;
    calculoCoste(500, 5000, 500, costeTabla, fichero_tabla);
    
    // Cerramos ficheros para asegurar que los datos se guardan en disco
    fichero_abb.close();
    fichero_map.close();
    fichero_tabla.close();
    
    cout << "Archivos generados: abb.dat, map.dat, tabla.dat" << endl;
    
    return 0;
}

/**
 * @brief Ejecuta el bucle de pruebas para diferentes tamaños y calcula el tiempo medio.
 * @param ini Talla inicial.
 * @param fin Talla final.
 * @param inc Salto entre tallas.
 * @param f Función que realiza la inserción y búsqueda.
 * @param file Fichero donde guardar los resultados.
 */
void calculoCoste(unsigned ini, unsigned fin, int inc, void (*f)(unsigned), ostream& file)
{
    // Bucle principal: va aumentando el número de elementos (talla)
    for (unsigned talla = ini; talla <= fin; talla += inc)
    {
        clock_t t_ini = clock();
        
        f(talla);
        
        clock_t t_fin = clock();
        
        double tiempo = (double)(t_fin - t_ini) / CLOCKS_PER_SEC;
        
        double coste_medio = tiempo / (100 * talla);
        
        file << talla << " " << coste_medio << endl;
        
        cout << "  Talla " << talla << ": " << coste_medio << " segundos" << endl;
    }
}

/**
 * @brief Prueba de rendimiento para el Árbol Binario de Búsqueda (ABB).
 * @param n Número de elementos a insertar y buscar.
 */
void costeABB(unsigned n)
{
    ABB<unsigned, Alumno> arbol;
    vector<unsigned> dnis; // Vector auxiliar para recordar que claves hemos insertado
    
    // --- FASE 1: INSERCION ---
    for (unsigned i = 0; i < n; i++)
    {
        Alumno alum(1); // Generar alumno aleatorio
        
        //Extraer la parte numérica del DNI para usarla como clave (unsigned)
        string dni_str = alum.getDNI().substr(0, alum.getDNI().length()-1);
        unsigned dni_num = stoul(dni_str); 
        
        arbol.insertar(dni_num, alum);
        dnis.push_back(dni_num); // Guardamos la clave para buscarla luego
    }
    
    // --- FASE 2: BUSQUEDA ---
    Alumno resultado; 
    for (unsigned rep = 0; rep < 100; rep++)
    {
        for (unsigned i = 0; i < n; i++)
        {
            arbol.buscar(dnis[i], resultado); 
        }
    }
}

/**
 * @brief Prueba de rendimiento para std::map (Árbol Rojo-Negro).
 * @param n Número de elementos a insertar y buscar.
 */
void costeMap(unsigned n)
{
    map<unsigned, Alumno> mapa;
    vector<unsigned> dnis;
    
    // --- FASE 1: INSERCION ---
    for (unsigned i = 0; i < n; i++)
    {
        Alumno alum(1);
        string dni_str = alum.getDNI().substr(0, alum.getDNI().length()-1);
        unsigned dni_num = stoul(dni_str);
        
        mapa[dni_num] = alum;
        dnis.push_back(dni_num);
    }
    
    // --- FASE 2: BUSQUEDA ---
    for (unsigned rep = 0; rep < 100; rep++)
    {
        for (unsigned i = 0; i < n; i++)
        {
            
            mapa.find(dnis[i]); 
        }
    }
}

/**
 * @brief Prueba de rendimiento para la Tabla Hash.
 * @param n Número de elementos a insertar y buscar.
 */
void costeTabla(unsigned n)
{
    Tabla<unsigned, Alumno> tabla(2 * n); 
    vector<unsigned> dnis;
    Alumno resultado;
    
    // --- FASE 1: INSERCION ---
    for (unsigned i = 0; i < n; i++)
    {
        Alumno alum(1);
        string dni_str = alum.getDNI().substr(0, alum.getDNI().length()-1);
        unsigned dni_num = stoul(dni_str);
        
        tabla.insertar(dni_num, alum);
        dnis.push_back(dni_num);
    }
    
    // --- FASE 2: BuSQUEDA ---
    for (unsigned rep = 0; rep < 100; rep++)
    {
        for (unsigned i = 0; i < n; i++)
        {
            tabla.buscar(dnis[i], resultado); 
        }
    }
}

/*
 ==============================================================================
 CONCLUSIONES DEL ANALISIS DE COSTES
 ==============================================================================

 1. TABLA HASH (Rendimiento Superior):
    - Es la estructura MAS RAPIDA de las tres comparadas.
    - Su curva es plana, lo que demuestra un coste O(1) (Tiempo Constante).
    - El tiempo de busqueda se mantiene estable y no aumenta aunque
      el numero de elementos (n) crezca de 500 a 5000.

 2. ABB y MAP (Estructuras de Arbol):
    - Son visiblemente mas lentas que la Tabla Hash.
    - Muestran una tendencia CRECIENTE.
    - Esto confirma un coste O(log n) (Tiempo Logaritmico): a medida que
      insertamos mas elementos, el tiempo de busqueda aumenta.

 3. RESUMEN FINAL:
    - Los datos confirman la teoria: para busquedas por clave exacta,
      la Tabla Hash es mucho mas eficiente que los arboles (ABB o Map),
      especialmente cuando trabajamos con un gran volumen de datos.
 ==============================================================================
*/