#pragma once
#include <stdexcept>
using namespace std;

/**
 * @class PilaVacia
 * @brief Excepción lanzada cuando se intenta acceder o desapilar una pila vacía.
 */
class PilaVacia : public logic_error
{
public:
    /**
     * @brief Constructor de la excepción.
     */
    PilaVacia() : logic_error("Error: la pila esta vacia") {}
};

/**
 * @class PilaLlena
 * @brief Excepción lanzada cuando se intenta apilar un elemento en una pila llena.
 */
class PilaLlena : public logic_error
{
public:
    /**
     * @brief Constructor de la excepción.
     */
    PilaLlena() : logic_error("Error: la pila esta llena.") {}
};
