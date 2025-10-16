    /**
 * Practica 1.
 *
 * Practicas Estructuras de Datos y Algoritmos
 * Curso 2021 - 2022
 *
 * (c) Departament d'Informatica.
 *     Universitat de Valencia
 *
 * @file complejo.h
 */
 
#ifndef _COMPLEJO_H
#define _COMPLEJO_H

#include <cmath>
using namespace std;

/**
 * Complejo.
 * Representacion de numeros complejos en notacion polar
 * 
 */
class Complejo
{
private:
    /**
     * Modulo del complejo
     */
    float mod;

    /**
     * Argumento del complejo.
     * Expresado en radianes.
     */
    float arg;

public:

    /**
    * Crea un complejo inicializandolo a 0 + 0i
    */
    Complejo() {mod = 0; arg = 0; };

    /**
    * Crea un complejo con un valor inicial para el modulo y el argumento.
    * @param m Modulo.
    * @param a Argumento.
    */
    Complejo(float m, float a);

    /**
    * Invariante de clase para complejos en notacion polar.
    * @return Indicador de complejo valido (true) o invalido (false).
    */
    bool Inv() const;

    /**
    * Obtiene la parte real del complejo.
    * @return Parte real del complejo.
    */
    float GetReal() const;

    /**
    * Obtiene la parte imaginaria del complejo.
    * @return Parte imaginaria del complejo.
    */
    float GetImag() const;

    /**
    * Suma dos complejos.
    * @param c Parametro explicito de la suma.
    * @return Complejo resultado de la suma.
    */
    Complejo operator+(const Complejo c) const;

    /**
    * Multiplica dos complejos.
    * @param c Parametro explicito del producto.
    * @return Complejo resultado del producto.
    */
    Complejo operator*(const Complejo &c) const;

    /**
    * Compara dos complejos e indica si son iguales.
    * @param c Parametro explicito de la comparacion.
    * @return Indicador de igualdad.
    */
    bool operator==(const Complejo c) const;

    /**
    * Asigna nuevos valores al modulo y el argumento del complejo.
    * @param m Nuevo modulo.
    * @param a Nuevo argumento.
    */
    void Set(float m, float a);

    /**
    * Asigna nuevos valores al modulo y argumento del complejo
    * a partir de nuevas coordenadas polares para el mismo.
    *
    * @param r Coordenada real
    * @param i Coordenada imaginaria
    */
    void SetRec(float r, float i);

    /**
    * Imprime el complejo por pantalla
    */
    void Print() const;
};

#endif
