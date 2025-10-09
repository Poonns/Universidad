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

#include <cstdlib>
#include <iostream>
#include <cassert>

#include "complejo.h"

using namespace std;

/** 
 * Diferencia maxima que se considera cero al comparar dos reales
 */
const float CERO = 0.0001;

Complejo::Complejo(float m, float a) 
{
   mod = m;
   arg = a;
} 

bool Complejo::operator==(const Complejo c) const 
{
    if(mod == 0 && c.mod == 0)
        return true;
    else
        return (fabs(mod - c.mod) < CERO && fabs(arg - c.arg) < CERO);
}

void Complejo::setRec(float re, float im) 
{
    mod = sqrt(re*re + im*im);
    if(re == 0 && im == 0)
        arg = 0;
    else
        arg = atan2(im, re);
    if(arg < 0)
        arg = arg + 2 * M_PI;

    assert(inv());
}

float Complejo::getReal() const 
{
    return mod * cos(arg);
}

float Complejo::getImag() const 
{
    return mod * sin(arg);
}

bool Complejo::inv() const{
    if((mod >= 0) && (arg >= 0) && (arg < 2*M_PI))
        return true;
    return false;
}

void Complejo::set(float m, float a) const{
    if(mod < 0){
        mod = -m;
        a += 2*M_PI;
    } else {
        mod = m;
    }

    if(a < 0){
        while(a < 0){
            a += 2*M_PI;
        }
    } else {
        while(a >= 2*M_PI){
            a -= 2*M_PI;
        }
    }

    arg = a;

    assert(inv());
}

Complejo& Complejo::operator+(const Complejo c) const{
    Complejo nuevo;
    
    float r = getReal() + c.getReal();
    float i = getImag() + c.getImag();

    nuevo.setRec(r, i);

    return nuevo;
}

Complejo Complejo::operator*(const Complejo &c) const {
    Complejo result;

    result.set(mod * c.mod, arg + c.arg); 

    return result;
}

void Complejo::print() const {
    std::cout << getReal() << " + " << getImag() << "i" << std::endl;
}





