/**
 * Arbol Binario de Búsqueda (ABB)
 * Curso 2025/2026
 * @file abb.h
 */

#ifndef _ABB_H
#define _ABB_H

#include<iostream>

using namespace std;

template<typename TipoClave, typename TipoDato>
class ABB
{
public:
    ABB(); 
    ~ABB();
    bool buscar(const TipoClave &, TipoDato&) const; 
    bool insertar(const TipoClave &, const TipoDato&); 
    bool vacio() const;
    void inorden(ostream & sal) const;
    unsigned altura() const;
    unsigned tam() const;
    double mediaDNI() const;
    unsigned cuentaDNI8() const;
private:
    class NodoBin
    {
    public:
        TipoClave clave; 
        TipoDato  dato;
        NodoBin * izq;
        NodoBin * der;

        NodoBin(): izq(NULL), der(NULL) {};
        NodoBin(TipoClave c, TipoDato d)
        : clave(c), dato(d), izq(NULL), der(NULL) {};
        ~NodoBin();
    };

    typedef NodoBin * PtrNodoBin;

    bool buscarAux(PtrNodoBin, const TipoClave &, TipoDato&) const; 
    bool insertarAux(PtrNodoBin &, const TipoClave &, const TipoDato&); 
    void inordenAux(ostream & sal, PtrNodoBin) const;
    // Incluir aquí resto de funciones auxiliares
    unsigned alturaAux(PtrNodoBin) const;                                    
    unsigned tamAux(PtrNodoBin) const;                                       
    void mediaDNIAux(PtrNodoBin, unsigned&, unsigned long long&) const;     
    unsigned cuentaDNI8Aux(PtrNodoBin) const;

    PtrNodoBin miRaiz;
};


//Implementacion de las funciones

//Constructor
template<typename TipoClave, typename TipoDato>
ABB<TipoClave, TipoDato>::ABB()
{
    miRaiz = NULL;
}

//Destructor de NodoBin (recursivo)
template<typename TipoClave, typename TipoDato>
ABB<TipoClave, TipoDato>::NodoBin::~NodoBin()
{
    if (izq != NULL) 
		delete izq;
    if (der != NULL) 
		delete der;
}

//Destructor de ABB
template<typename TipoClave, typename TipoDato>
ABB<TipoClave, TipoDato>::~ABB()
{
    if (miRaiz != NULL)
        delete miRaiz;
}

//Metodo vacio
template<typename TipoClave, typename TipoDato>
bool ABB<TipoClave, TipoDato>::vacio() const
{
    return (miRaiz == NULL);
}

//Metodo Insertar 
template<typename TipoClave, typename TipoDato>
bool ABB<TipoClave, TipoDato>::insertar(const TipoClave & c, const TipoDato & d)
{
    return insertarAux(miRaiz, c, d);
}

//Metodo Insertar (auxiliar recursivo)
template<typename TipoClave, typename TipoDato>
bool ABB<TipoClave, TipoDato>::insertarAux(PtrNodoBin & nodo, const TipoClave & c, const TipoDato & d)
{
    bool resultado;
    
    if (nodo == NULL)
    {
        nodo = new NodoBin(c, d);
        resultado = true;
    }
    else if (c < nodo->clave)
    {
        resultado = insertarAux(nodo->izq, c, d);
    }
    else if (c > nodo->clave)
    {
        resultado = insertarAux(nodo->der, c, d);
    }
    else
    {
        resultado = false; //clave duplicada
    }
    
    return resultado;
}

//Metodo Buscar (público)
template<typename TipoClave, typename TipoDato>
bool ABB<TipoClave, TipoDato>::buscar(const TipoClave & c, TipoDato & d) const
{
    return buscarAux(miRaiz, c, d);
}

//Metodo Buscar (auxiliar recursivo)
template<typename TipoClave, typename TipoDato>
bool ABB<TipoClave, TipoDato>::buscarAux(PtrNodoBin nodo, const TipoClave & c, TipoDato & d) const
{
    bool encontrado;
    
    if (nodo != NULL)
    {
        if (c == nodo->clave)
        {
            d = nodo->dato;
            encontrado = true;
        }
        else if (c < nodo->clave)
        {
            encontrado = buscarAux(nodo->izq, c, d);
        }
        else
        {
            encontrado = buscarAux(nodo->der, c, d);
        }
    }
    
    return encontrado;
}

//Metodo Inorden
template<typename TipoClave, typename TipoDato>
void ABB<TipoClave, TipoDato>::inorden(ostream & sal) const
{
    inordenAux(sal, miRaiz);
}

//Metodo Inorden (auxiliar recursivo)
template<typename TipoClave, typename TipoDato>
void ABB<TipoClave, TipoDato>::inordenAux(ostream & sal, PtrNodoBin nodo) const
{
    if (nodo != NULL)
    {
        inordenAux(sal, nodo->izq);
        sal << nodo->dato;
        inordenAux(sal, nodo->der);
    }
}

//Metodo tam 
template<typename TipoClave, typename TipoDato>
unsigned ABB<TipoClave, TipoDato>::tam() const
{
    return tamAux(miRaiz);
}

//Metodo tam (auxiliar recursivo)
template<typename TipoClave, typename TipoDato>
unsigned ABB<TipoClave, TipoDato>::tamAux(PtrNodoBin nodo) const
{
    unsigned resultado = 0;
    
    if (nodo != NULL)
    {
        resultado = 1 + tamAux(nodo->izq) + tamAux(nodo->der);
    }
    
    return resultado;
}

//Metodo Altura 
template<typename TipoClave, typename TipoDato>
unsigned ABB<TipoClave, TipoDato>::altura() const
{
    return alturaAux(miRaiz);
}

//Metodo Altura (auxiliar recursivo)
template<typename TipoClave, typename TipoDato>
unsigned ABB<TipoClave, TipoDato>::alturaAux(PtrNodoBin nodo) const
{
    unsigned resultado = 0;
    unsigned altIzq;
    unsigned altDer;
    
    if (nodo != NULL)
    {
        altIzq = alturaAux(nodo->izq);
        altDer = alturaAux(nodo->der);
        resultado = 1 + max(altIzq, altDer);
    }
    
    return resultado;
}

//Metodo media DNI 
template<typename TipoClave, typename TipoDato>
double ABB<TipoClave, TipoDato>::mediaDNI() const
{
    unsigned count = 0;
    unsigned long long sum = 0;
    double resultado = 0.0;
    
    mediaDNIAux(miRaiz, count, sum);
    
    if (count > 0)
    {
        resultado = sum / count;
    }
    
    return resultado;
}

//Metodo media DNI (auxiliar recursivo)
template<typename TipoClave, typename TipoDato>
void ABB<TipoClave, TipoDato>::mediaDNIAux(PtrNodoBin nodo, unsigned & count, unsigned long long & sum) const
{
    if (nodo != NULL)
    {
        mediaDNIAux(nodo->izq, count, sum);
        sum += nodo->clave;
        count++;
        mediaDNIAux(nodo->der, count, sum);
    }
}

//Metodo Cuenta DNI que empiezan por 8
template<typename TipoClave, typename TipoDato>
unsigned ABB<TipoClave, TipoDato>::cuentaDNI8() const
{
    return cuentaDNI8Aux(miRaiz);
}

//Metodo Cuenta DNI que empiezan por 8 (auxiliar recursivo)
template<typename TipoClave, typename TipoDato>
unsigned ABB<TipoClave, TipoDato>::cuentaDNI8Aux(PtrNodoBin nodo) const
{
    unsigned count = 0;
    
    if (nodo != NULL)
    {
        // Si el DNI está en el rango [80000000, 89999999]
        if (nodo->clave >= 80000000 && nodo->clave <= 89999999)
        {
            count = 1;
        }
        
        // Solo exploramos izquierda si puede haber DNIs >= 80000000
        if (nodo->clave >= 80000000)
        {
            count += cuentaDNI8Aux(nodo->izq);
        }
        
        // Solo exploramos derecha si puede haber DNIs <= 89999999
        if (nodo->clave <= 89999999)
        {
            count += cuentaDNI8Aux(nodo->der);
        }
    }
    
    return count;
}

#endif
