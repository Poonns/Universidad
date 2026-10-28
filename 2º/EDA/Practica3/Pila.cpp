#include "Pila.h"
#include <iostream>
using namespace std;

/**
 * @brief Constructor de la clase Pila.
 * @param max Tamaño máximo de la pila.
 */
Pila::Pila(unsigned max)
{
    this->max = max;
    v = new int[max];
    ind_cima = 0;
    accesos = 0;
}

/**
 * @brief Constructor de copia.
 * @param p Objeto Pila a copiar.
 */
Pila::Pila(const Pila& p)
{
    max = p.max;
    ind_cima = p.ind_cima;
    accesos = p.accesos;
    v = new int[max];
    for (unsigned i = 0; i < ind_cima; i++)
        v[i] = p.v[i];
}

/**
 * @brief Destructor de la clase Pila.
 */
Pila::~Pila()
{
    delete[] v;
}

/**
 * @brief Operador de asignación.
 * @param q Pila a asignar.
 * @return Referencia a la pila actual.
 */
const Pila& Pila::operator=(const Pila& q)
{
    if (this != &q)
    {
        delete[] v;
        max = q.max;
        ind_cima = q.ind_cima;
        accesos = q.accesos;
        v = new int[max];
        for (unsigned i = 0; i < ind_cima; i++)
            v[i] = q.v[i];
    }
    return *this;
}

/**
 * @brief Inserta un elemento en la pila.
 * @param valor Valor entero a apilar.
 * @throw PilaLlena Si la pila está llena.
 */
void Pila::Apilar(int valor)
{
    if (ind_cima >= max)
        throw PilaLlena();
    v[ind_cima++] = valor;
    accesos++;
}

/**
 * @brief Elimina el elemento superior de la pila.
 * @throw PilaVacia Si la pila está vacía.
 */
void Pila::Desapilar()
{
    if (ind_cima == 0)
        throw PilaVacia();
    ind_cima--;
    accesos++;
}

/**
 * @brief Devuelve el elemento superior de la pila.
 * @return Valor entero de la cima.
 * @throw PilaVacia Si la pila está vacía.
 */
int Pila::Cima()
{
    if (ind_cima == 0)
        throw PilaVacia();
    accesos++;
    return v[ind_cima - 1];
}

/**
 * @brief Indica si la pila está vacía.
 * @return 1 si la pila está vacía, 0 en caso contrario.
 */
int Pila::Vacia()
{
    return ind_cima == 0;
}

/**
 * @brief Devuelve el número de accesos realizados.
 * @return Número de accesos.
 */
unsigned Pila::getAccesos()
{
    return accesos;
}

/**
 * @brief Reinicia el contador de accesos a cero.
 */
void Pila::resetAccesos()
{
    accesos = 0;
}

/**
 * @brief Muestra el contenido de la pila por pantalla.
 */
void Pila::mostrar()
{
    cout << "Contenido de la pila (de cima a base): ";
    for (int i = ind_cima - 1; i >= 0; i--)
        cout << v[i] << " ";
    cout << endl;
}
