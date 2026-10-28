/********************test_arbol.cpp*************//**
 *
 * @file test_arbol.cpp
 *
 * @brief Practicas de Estructura de datos y algoritmos
 *
 * Programa que crea un arbol de la clase abb.h y
 * comprueba el funcionamiento de sus metodos
 *
 * @version 1.0
 *
 * @author Meryem Bagho
 * @author Hector Pons
 *
 * @date 11/11/2025
 *
*//***************************************************/

#include <iostream>
#include <cstdlib>
#include <ctime>
#include <string> 
#include "abb.h"
#include "alumno.h"

using namespace std;

// PROTOTIPO DE LA FUNCION 
unsigned DniANumero(const string &);

int main (void)
{
    srand(time(NULL));

    // Declaraciones
    string dni_comp, dni_sin_letra;
    unsigned clave;
    unsigned dni1;
    unsigned dni2;
    unsigned dni3;
    Alumno resultado1;
    Alumno resultado2;
    Alumno resultado3;

    ABB<unsigned, Alumno> arbol;

    cout << "Generacion de 1000 estudiantes de un solo curso..." << endl;

    for (int i = 0; i < 1000; i++)
    {
        Alumno estudiante(1);

        // Conversion del DNI a una a clave
        dni_comp = estudiante.getDNI();
        dni_sin_letra = dni_comp.substr(0,dni_comp.length()-1);
        clave = stoul(dni_sin_letra);

        // Insertamos la clave en el arbol
        arbol.insertar(clave, estudiante);
    }

    // Usaremos los DNIs fijos del primer archivo para la búsqueda
    dni1 = 45678912; // DNI 1 (test1)
    dni2 = 23456789; // DNI 2 (test2)
    dni3 = 99999999; // DNI 3 (test3 - no existe)

    cout << "Alumnos insertados correctamente" << endl;

    // Mostrar el arbol in orden
    cout << "===== ÁRBOL EN INORDEN =====" << endl;
    arbol.inorden(cout);
    cout << endl;

    // Buscar 3 DNIS
    cout << "===== BÚSQUEDA DE DNIs =====" << endl;

    // DNI 1 (test1)
    if (arbol.buscar(dni1, resultado1))
    {
        cout << "DNI " << dni1 << " ENCONTRADO:" << endl;
        cout << resultado1;
    }
    else
    {
        cout << "DNI " << dni1 << " NO encontrado." << endl;
    }
    cout << endl;

    // DNI 2 (test2)
    if (arbol.buscar(dni2, resultado2))
    {
        cout << "DNI " << dni2 << " ENCONTRADO:" << endl;
        cout << resultado2;
    }
    else
    {
        cout << "DNI " << dni2 << " NO encontrado." << endl;
    }
    cout << endl;

    // DNI 3 (test3)
    if (arbol.buscar(dni3, resultado3))
    {
        cout << "DNI " << dni3 << " ENCONTRADO:" << endl;
        cout << resultado3;
    }
    else
    {
        cout << "DNI " << dni3 << " NO encontrado." << endl;
    }
    cout << endl;

    // Caracteristicas del arbol
    cout << "===== ESTADÍSTICAS DEL ÁRBOL =====" << endl;
    cout << "Tamanyo (número de nodos): " << arbol.tam() << endl;
    cout << "Altura del árbol: " << arbol.altura() << endl;
    cout << endl;

    // Media y contador de DNIS que comienzan por 8
    cout << "===== ANÁLISIS DE DNIs =====" << endl;
    cout << "Media de los DNIs: " << arbol.mediaDNI() << endl;
    cout << "DNIs que comienzan por 8: " << arbol.cuentaDNI8() << endl;
    cout << endl;


    // EL arbol esta equilibrado o no?
    cout << "===== ANÁLISIS DE EQUILIBRIO =====" << endl;
    cout << "¿Está el árbol equilibrado?" << endl;
    cout << "Para un árbol con 1000 nodos:" << endl;
    cout << "  ° Altura actual: " << arbol.altura() << endl;
    cout << "  ° Altura teórica equilibrada = 10 (log2(1000))" << endl;

    // Condición de equilibrio del primer archivo
    if (arbol.altura() <= 15)
    {
        cout << "  El árbol está relativamente equilibrado." << endl;
    }
    else
    {
        cout << "  El árbol NO está equilibrado (demasiado alto)." << endl;
    }

    return 0;
}

/**
 * @brief Implementación de la conversión de DNI a número
 * @param dni String con el DNI completo (8 dígitos + letra)
 * @return unsigned Valor numérico del DNI sin la letra
 * @details Extrae la parte numérica del DNI eliminando el último carácter (letra)
 * y lo convierte a unsigned mediante stoul
 */
unsigned DniANumero(const string & dni)
{
    string dniNumerico = dni.substr(0, dni.length() - 1);
    
    return stoul(dniNumerico);
}
