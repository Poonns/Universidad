/**
 * Practica 1.
 *
 * Practicas Estructuras de Datos y Algoritmos
 * Curso 2021 - 2022
 *
 * (c) Departament d'Informatica.
 *     Universitat de Valencia
 *
 * @file complejo.cpp
 */

#include <cmath>
#include <iostream>
#include <cassert>

#include "complejo.h"

using namespace std;

/** 
 * Diferencia maxima que se considera cero al comparar dos reales
 */
const float CERO = 0.0001;

/**
 * @brief Constructor con valores iniciales
 * @param m Módulo del complejo
 * @param a Argumento en radianes
 */

Complejo::Complejo(float m, float a) 
{
   mod = m;
   arg = a;
} 

/**
 * @brief Compara si dos complejos son iguales
 * @param c Complejo con el que se compara
 * @return true si ambos complejos son equivalentes, false en caso contrario
 */

bool Complejo::operator==(const Complejo c) const 
{
    if(mod == 0 && c.mod == 0)
        return true;
    else
        return (fabs(mod - c.mod) < CERO && fabs(arg - c.arg) < CERO);
}

/**
 * @brief Establece el complejo a partir de sus componentes rectangulares
 * @param re Parte real
 * @param im Parte imaginaria
 */

void Complejo::SetRec(float re, float im) 
{
    mod = sqrt(re*re + im*im);
    if(re == 0 && im == 0)
        arg = 0;
    else
        arg = atan2(im, re);
    if(arg < 0)
        arg = arg + 2 * M_PI;

    assert(Inv());
}

/**
 * @brief Obtiene la parte real del complejo
 * @return Parte real
 */

float Complejo::GetReal() const 
{
    return mod * cos(arg);
}

/**
 * @brief Obtiene la parte imaginaria del complejo
 * @return Parte imaginaria
 */

float Complejo::GetImag() const 
{
    return mod * sin(arg);
}

/**
 * @brief Comprueba si el complejo cumple las condiciones del invariante
 * @return true si el complejo es válido, false en caso contrario
 */

bool Complejo::Inv() const
{
    if((mod >= 0) && (arg >= 0) && (arg < 2*M_PI))
        return true;
    return false;
}

/**
 * @brief Asigna nuevos valores al módulo y argumento
 * @param m Nuevo módulo
 * @param a Nuevo argumento (radianes)
 */

void Complejo::Set(float m, float a)
{
    if(mod < 0)
    {
        mod = -m;
        a += 2*M_PI;
    } 
    else
        mod = m;

    if(a < 0)
    {
        while(a < 0)
            a += 2*M_PI;
    } 
    else 
    {
        while(a >= 2*M_PI)
            a -= 2*M_PI;
    }
    
    arg = a;

    assert(Inv());
}

/**
 * @brief Suma dos complejos
 * @param c Complejo que se suma
 * @return Complejo resultado de la operación
 */

Complejo Complejo::operator+(const Complejo c) const
{
    Complejo nuevo;
    
    float r = GetReal() + c.GetReal();
    float i = GetImag() + c.GetImag();

    nuevo.SetRec(r, i);

    return nuevo;
}

/**
 * @brief Multiplica dos complejos
 * @param c Complejo que multiplica
 * @return Complejo resultado de la operación
 */

Complejo Complejo::operator*(const Complejo &c) const 
{
    Complejo result;

    result.Set(mod * c.mod, arg + c.arg); 

    return result;
}

/**
 * @brief Muestra el número complejo con la forma: a + bi.
 */

void Complejo::Print() const 
{
    cout << GetReal() << " + " << GetImag() << "i" << endl;
}





