#pragma once
#include "error.h"

/**
 * @class Pila
 * @brief Implementa una pila de enteros con operaciones básicas de apilado y desapilado.
 *
 * Esta clase gestiona una estructura de datos tipo pila (LIFO: Last In, First Out)
 * usando un vector dinámico. Incluye control de errores mediante excepciones.
 */
class Pila
{
public:
    /**
     * @brief Constructor de la pila.
     * @param max Tamaño máximo de la pila (por defecto 10).
     */
    Pila(unsigned max = 10);

    /**
     * @brief Constructor de copia.
     * @param p Pila a copiar.
     */
    Pila(const Pila& p);

    /**
     * @brief Destructor de la pila.
     */
    ~Pila();

    /**
     * @brief Operador de asignación.
     * @param q Pila a asignar.
     * @return Referencia a la pila actual.
     */
    const Pila& operator=(const Pila& q);

    /**
     * @brief Inserta un elemento en la cima de la pila.
     * @param valor Valor entero a apilar.
     * @throw PilaLlena Si la pila está llena.
     */
    void Apilar(int valor);

    /**
     * @brief Elimina el elemento superior de la pila.
     * @throw PilaVacia Si la pila está vacía.
     */
    void Desapilar();

    /**
     * @brief Devuelve el valor del elemento superior (cima).
     * @return Entero correspondiente a la cima.
     * @throw PilaVacia Si la pila está vacía.
     */
    int Cima();

    /**
     * @brief Indica si la pila está vacía.
     * @return 1 si la pila está vacía, 0 en caso contrario.
     */
    int Vacia();

    /**
     * @brief Obtiene el número de accesos realizados a la pila.
     * @return Número de accesos acumulados.
     */
    unsigned getAccesos();

    /**
     * @brief Reinicia el contador de accesos a cero.
     */
    void resetAccesos();

    /**
     * @brief Muestra el contenido de la pila por pantalla.
     *
     * Los elementos se muestran desde la cima hasta la base.
     */
    void mostrar();

private:
    int* v; ///< Vector dinámico que almacena los elementos de la pila.
    unsigned ind_cima; ///< Índice del elemento superior.
    unsigned max; ///< Capacidad máxima de la pila.
    unsigned accesos; ///< Contador de accesos realizados.
};
