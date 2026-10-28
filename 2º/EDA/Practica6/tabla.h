/**
 * @file tabla.h
 * @brief Definición e implementación de una Tabla Hash con resolución de colisiones mediante encadenamiento.
 * @date Curso 2023/2024
 */

#ifndef _TABLA_H
#define _TABLA_H

#include <string>
#include <vector>
#include <list>
#include <iostream>
#include "alumno.h"

using namespace std;

/**
 * @class Tabla
 * @brief Clase plantilla para gestionar una Tabla Hash.
 * * Utiliza un vector de listas (std::vector<std::list>) para manejar las colisiones
 * (técnica de encadenamiento separado).
 * * @tparam TipoClave Tipo de datos para la clave (debe soportar la operación módulo %).
 * @tparam TipoDato Tipo de datos para el valor asociado a la clave.
 */
template<typename TipoClave, typename TipoDato>
class Tabla
{
public:
    /**
     * @struct ParDato
     * @brief Estructura auxiliar para almacenar el par (clave, valor) en las listas.
     */
    struct ParDato
    {
        TipoClave clave; ///< La clave única del elemento.
        TipoDato dato;   ///< El valor o dato asociado.
    };

    /**
     * @brief Constructor de la Tabla Hash.
     * @param tam Número de "buckets" o posiciones iniciales del vector.
     */
    Tabla(unsigned tam); 

    /**
     * @brief Busca un elemento en la tabla dado su clave.
     * @param elem La clave del elemento a buscar.
     * @param valor [out] Referencia donde se guardará el dato si se encuentra.
     * @return true si se encontró el elemento, false en caso contrario.
     */
    bool buscar(TipoClave elem, TipoDato& valor); 

    /**
     * @brief Inserta un nuevo elemento en la tabla.
     * * Si se produce una colisión, el elemento se añade al final de la lista correspondiente.
     * @param elem La clave del nuevo elemento.
     * @param valor El dato asociado a guardar.
     */
    void insertar(TipoClave elem, const TipoDato& valor); 

    /**
     * @brief Función hash simple.
     * @param clave La clave a transformar.
     * @return El índice calculado dentro del rango del vector.
     */
    unsigned hash(TipoClave clave) const; 

    /**
     * @brief Muestra el contenido de la tabla hash por un flujo de salida.
     * * Imprime cada posición del vector y la lista de elementos que contiene.
     * @param sal Flujo de salida (por ejemplo, cout).
     */
    void mostrar(ostream & sal) const; 

    /**
     * @brief Calcula el factor de carga de la tabla.
     * * El factor de carga es la relación entre el número total de elementos y el tamaño de la tabla.
     * @return Float con el factor de carga promedio.
     */
    float factorCarga() const;

    /**
     * @brief Cuenta cuántas posiciones (buckets) tienen exactamente 'n' elementos.
     * @param n El número de elementos (longitud de la lista) a buscar.
     * @return Número de posiciones que cumplen la condición.
     */
    unsigned num_posN(int n) const;

private:
    typedef list<ParDato> ListaDatos; ///< Alias para la lista de colisiones.
    vector<ListaDatos> t;             ///< Vector que contiene las listas de datos (Buckets).
};

// --------------------------------------------------------------------
// IMPLEMENTACIÓN DE LOS MÉTODOS
// --------------------------------------------------------------------

/**
 * @brief Constructor de la clase.
 * Inicializa el tamaño del vector de listas.
 * @param tam Tamaño inicial de la tabla hash.
 */
template<typename TipoClave, typename TipoDato>
Tabla<TipoClave, TipoDato>::Tabla(unsigned tam)
{
    t.resize(tam);
}

/**
 * @brief Busca un elemento en la tabla.
 * @param elem Clave del elemento a buscar.
 * @param valor [out] Referencia donde se devuelve el dato asociado si se encuentra.
 * @return true si se encuentra la clave, false si no.
 */
template<typename TipoClave, typename TipoDato>
bool Tabla<TipoClave, TipoDato>::buscar(TipoClave elem, TipoDato& valor)
{
    unsigned i = hash(elem);
    
    for(auto it = t[i].begin(); it != t[i].end(); ++it)
    {
        if(it->clave == elem)
        {
            valor = it->dato;
            return true;
        }
    }
    return false;
}

/**
 * @brief Inserta un nuevo elemento en la tabla hash.
 * @param elem Clave del elemento.
 * @param valor Dato asociado a la clave.
 */
template<typename TipoClave, typename TipoDato>
void Tabla<TipoClave, TipoDato>::insertar(TipoClave elem, const TipoDato & valor)
{
    unsigned i = hash(elem);
    t[i].push_back( ParDato{elem, valor} );
}

/**
 * @brief Función hash para calcular el índice.
 * @param clave Clave a procesar.
 * @return Índice correspondiente dentro del vector.
 */
template<typename TipoClave, typename TipoDato>
unsigned Tabla<TipoClave, TipoDato>::hash(TipoClave clave) const
{
    return clave % t.size(); 
}

/**
 * @brief Muestra el contenido de la tabla por un flujo de salida.
 * @param sal Flujo de salida (ej. std::cout).
 */
template<typename TipoClave, typename TipoDato>
void Tabla<TipoClave, TipoDato>::mostrar(ostream & sal) const
{
    for (unsigned i = 0; i < t.size(); i++)
    {
        sal << i << ": ";
        for (auto it = t[i].begin(); it != t[i].end(); ++it)
            sal << "(" << it->clave << ", " << it->dato << ")";
        sal << endl; 
    }
}

/**
 * @brief Calcula el factor de carga de la tabla (elementos / tamaño).
 * @return El factor de carga como float.
 */
template<typename TipoClave, typename TipoDato>
float Tabla<TipoClave, TipoDato>::factorCarga() const
{
    unsigned total_elementos = 0; 
    for (unsigned i = 0; i < t.size(); i++)
    {
        total_elementos += t[i].size();
    }
    return (float)total_elementos / t.size();
}

/**
 * @brief Devuelve el número de posiciones que contienen exactamente n elementos.
 * @param n Cantidad de elementos a buscar por posición.
 * @return Número de posiciones que cumplen la condición.
 */
template<typename TipoClave, typename TipoDato>
unsigned Tabla<TipoClave, TipoDato>::num_posN(int n) const
{
    unsigned contador = 0;
    for (unsigned i = 0; i < t.size(); ++i) 
    {
        if (t[i].size() == (unsigned)n) 
        {
            contador++;
        }
    }
    return contador;
}

#endif