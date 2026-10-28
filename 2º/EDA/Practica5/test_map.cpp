/**
 * @file test_map.cpp
 * @brief Programa de prueba para la implementación usando std::map de C++
 * @details Este programa genera 1000 alumnos aleatorios, los inserta en un map usando
 * el DNI como clave, y realiza diversas operaciones de búsqueda y estadísticas.
 * Demuestra el uso de contenedores estándar de la STL como alternativa al ABB.
 */

#include <iostream>
#include <cstdlib>
#include <ctime>
#include <map>
#include "alumno.h"

using namespace std; 

// PROTOTIPO DE LA FUNCION
unsigned DniANumero(const string &);

/**
 * @brief Función principal del programa
 * @details Realiza las siguientes operaciones usando std::map:
 *          1. Genera 1000 alumnos aleatorios y los inserta en el map
 *          2. Muestra el map mediante recorrido inorden (automático)
 *          3. Busca 3 DNIs (2 existentes y 1 no existente)
 *          4. Muestra el tamaño del map
 *          5. Calcula la media de los DNIs y muestra los que empiezan por 8
 * @return int Código de retorno del programa (0 = éxito)
 */
int main (void)
{
    srand(time(NULL));

    // Declaracion de variables
    Alumno alumno; ///< Objeto alumno para crear instancias
    unsigned dni1, dni2, dni3; ///< DNIs para realizar búsquedas de prueba
    map<unsigned, Alumno> mapaAlumno; ///< Map de la STL que almacena alumnos con DNI como clave

    // Rellenar el map con 1000 alumnos
    cout << "Generando los alumnos..." << endl;
    for (int i = 0; i < 1000; i++)
    {
        alumno = Alumno(rand() % 4 + 1);
        unsigned claveDNI = DniANumero(alumno.getDNI());

        // Guardar los primeros 2 DNI para buscarlos despues
        if (i == 0)
            dni1 = claveDNI;
        if(i == 1)
            dni2 = claveDNI;

        mapaAlumno[claveDNI] = alumno;
    }
    dni3 = 9999999; ///< DNI que no existe en el map (para prueba de búsqueda fallida)

    // 1. Mostrar el map en pantalla (recorrido inorden automatico)
    cout << "1. Recorrido inorden del map: " << endl;
    
    for (map<unsigned, Alumno>::iterator it = mapaAlumno.begin(); it != mapaAlumno.end(); ++it)
    {
        cout << it->second << endl; ///< it->second es el alumno
    }
    cout << endl;

    // 2. Buscar 3 DNI y 1 que no existe
    cout << "2. Busqueda de DNI: " << endl;
    
    // DNI1 - Búsqueda del primer DNI guardado
    cout << "Buscando DNI: "  << dni1 << endl;
    map<unsigned, Alumno>::iterator it1 = mapaAlumno.find(dni1);
    if (it1 != mapaAlumno.end())
    {
        cout << "DNI encontrado: " << endl;
        cout << it1->second;
    }
    else
        cout << "DNI no encontrado" << endl;
    cout << endl;

    // DNI2 - Búsqueda del segundo DNI guardado
    cout << "Buscando DNI: "  << dni2 << endl;
    map<unsigned, Alumno>::iterator it2 = mapaAlumno.find(dni2);
    if (it2 != mapaAlumno.end())
    {
        cout << "DNI encontrado: " << endl;
        cout << it2->second << endl;
    }
    else
        cout << "DNI no encontrado" << endl;
    cout << endl;

    // DNI3 - Búsqueda de un DNI que no existe
    cout << "Buscando DNI: "  << dni3 << endl;
    map<unsigned, Alumno>::iterator it3 = mapaAlumno.find(dni3);
    if (it3 != mapaAlumno.end())
    {
        cout << "DNI encontrado: " << endl;
        cout << it3->second;
    }
    else
        cout << "DNI no encontrado" << endl;

    // 3. Mostrar el tamanyo (el map no expone la altura)
    cout << "El tamanyo del map: " << mapaAlumno.size() << endl;
    cout << endl;

    // 4. Media de los DNI y cuantos DNI comienzan por 8
    
    /**
     * Calcular la media de todos los DNIs almacenados
     * Se recorre el map sumando todas las claves (DNIs)
     */
    double suma = 0; 
    for (map<unsigned, Alumno>::iterator it = mapaAlumno.begin(); it != mapaAlumno.end(); ++it)
        suma += it->first;

    double media = suma / mapaAlumno.size();
    cout << "Media de los DNI: " << media << endl;
    cout << endl;

    /**
     * Contar y mostrar todos los DNI que comienzan por 8
     * Usando lower_bound y upper_bound para búsqueda eficiente en rango
     */
    cout << "Los DNI que comienzan por 8 (80000000 - 89999999): " << endl;

    // Mostrar todos los DNI en ese rango
    int contador = 0; 
    for (map<unsigned, Alumno>::iterator it = mapaAlumno.begin(); it != mapaAlumno.end(); ++it)
    {
        unsigned dni = it->first;
        // Comprobamos si el DNI está en el rango [
        if (dni >= 80000000 && dni <= 89999999)
        {
            cout << it->second << endl;
            contador++;
        }
    }

    cout << "Total de DNI que comienzan por 8: " << contador << endl;

    return 0;
}

/**
 * @brief Implementación de la conversión de DNI a número
 * @param dni String con el DNI completo (8 dígitos + letra)
 * @return unsigned Valor numérico del DNI sin la letra
 * @details Extrae la parte numérica del DNI eliminando el último carácter (letra)
 *          y lo convierte a unsigned mediante stoul
 */
unsigned DniANumero(const string & dni)
{
    string dniNumerico = dni.substr(0, dni.length() - 1);
    return stoul(dniNumerico);
}
