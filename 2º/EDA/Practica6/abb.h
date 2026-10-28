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


// --------------------------------------------------------------------
// IMPLEMENTACIÓN DE LOS MÉTODOS
// --------------------------------------------------------------------

/**
 * @brief Constructor por defecto.
 * Inicializa el árbol con la raíz a NULL (vacío).
 */
template<typename TipoClave, typename TipoDato>
ABB<TipoClave, TipoDato>::ABB()
{
    miRaiz = NULL;
}

/**
 * @brief Destructor de la clase nodo (Recursivo).
 * Libera la memoria de los hijos izquierdo y derecho antes de eliminarse a sí mismo.
 */
template<typename TipoClave, typename TipoDato>
ABB<TipoClave, TipoDato>::NodoBin::~NodoBin()
{
    if (izq != NULL) 
		delete izq;
    if (der != NULL) 
		delete der;
}

/**
 * @brief Destructor de la clase ABB.
 * Invoca el borrado recursivo comenzando por la raíz.
 */
template<typename TipoClave, typename TipoDato>
ABB<TipoClave, TipoDato>::~ABB()
{
    if (miRaiz != NULL)
        delete miRaiz;
}

/**
 * @brief Comprueba si el árbol está vacío.
 * @return true si la raíz es NULL, false en caso contrario.
 */
template<typename TipoClave, typename TipoDato>
bool ABB<TipoClave, TipoDato>::vacio() const
{
    return (miRaiz == NULL);
}

/**
 * @brief Inserta un nuevo elemento en el árbol (Método público).
 * @param c Clave del elemento.
 * @param d Dato asociado.
 * @return true si se insertó correctamente, false si la clave ya existía.
 */
template<typename TipoClave, typename TipoDato>
bool ABB<TipoClave, TipoDato>::insertar(const TipoClave & c, const TipoDato & d)
{
    return insertarAux(miRaiz, c, d);
}

/**
 * @brief Función auxiliar recursiva para la inserción.
 * @param nodo Referencia al puntero del nodo actual.
 * @param c Clave a insertar.
 * @param d Dato a insertar.
 * @return true si se inserta, false si hay duplicados.
 */
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
        resultado = false; // clave duplicada
    }
    
    return resultado;
}

/**
 * @brief Busca un dato asociado a una clave (Método público).
 * @param c Clave a buscar.
 * @param d [out] Referencia donde se guardará el dato si se encuentra.
 * @return true si la clave existe, false si no.
 */
template<typename TipoClave, typename TipoDato>
bool ABB<TipoClave, TipoDato>::buscar(const TipoClave & c, TipoDato & d) const
{
    return buscarAux(miRaiz, c, d);
}

/**
 * @brief Función auxiliar recursiva de búsqueda.
 * @param nodo Puntero al nodo actual.
 * @param c Clave buscada.
 * @param d [out] Variable para devolver el dato.
 * @return true si encontrado, false si no.
 */
template<typename TipoClave, typename TipoDato>
bool ABB<TipoClave, TipoDato>::buscarAux(PtrNodoBin nodo, const TipoClave & c, TipoDato & d) const
{
    bool encontrado = false; // Inicializar para evitar retornos indefinidos
    
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

/**
 * @brief Realiza el recorrido inorden del árbol (Método público).
 * @param sal Flujo de salida.
 */
template<typename TipoClave, typename TipoDato>
void ABB<TipoClave, TipoDato>::inorden(ostream & sal) const
{
    inordenAux(sal, miRaiz);
}

/**
 * @brief Auxiliar recursivo para recorrido inorden (I-R-D).
 * @param sal Flujo de salida.
 * @param nodo Nodo actual.
 */
template<typename TipoClave, typename TipoDato>
void ABB<TipoClave, TipoDato>::inordenAux(ostream & sal, PtrNodoBin nodo) const
{
    if (nodo != NULL)
    {
        inordenAux(sal, nodo->izq);
        sal << nodo->dato << " "; // Se suele añadir un espacio o salto
        inordenAux(sal, nodo->der);
    }
}

/**
 * @brief Devuelve el número de nodos del árbol (Método público).
 * @return Tamaño del árbol.
 */
template<typename TipoClave, typename TipoDato>
unsigned ABB<TipoClave, TipoDato>::tam() const
{
    return tamAux(miRaiz);
}

/**
 * @brief Auxiliar recursivo para calcular el tamaño.
 * @param nodo Nodo actual.
 * @return Suma de 1 (actual) + tamaño hijo izq + tamaño hijo der.
 */
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

/**
 * @brief Calcula la altura del árbol (Método público).
 * @return Altura del árbol.
 */
template<typename TipoClave, typename TipoDato>
unsigned ABB<TipoClave, TipoDato>::altura() const
{
    return alturaAux(miRaiz);
}

/**
 * @brief Auxiliar recursivo para calcular la altura.
 * @param nodo Nodo actual.
 * @return 1 + la altura máxima entre los hijos.
 */
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
        // Nota: max requiere #include <algorithm> o <utility>
        resultado = 1 + max(altIzq, altDer);
    }
    
    return resultado;
}

/**
 * @brief Calcula la media de los valores de las claves (DNIs).
 * @return Valor de la media como double.
 */
template<typename TipoClave, typename TipoDato>
double ABB<TipoClave, TipoDato>::mediaDNI() const
{
    unsigned count = 0;
    unsigned long long sum = 0;
    double resultado = 0.0;
    
    mediaDNIAux(miRaiz, count, sum);
    
    if (count > 0)
    {
        resultado = (double)sum / count;
    }
    
    return resultado;
}

/**
 * @brief Auxiliar recursivo para calcular suma y cantidad de nodos.
 * @param nodo Nodo actual.
 * @param count [in/out] Contador de nodos visitados.
 * @param sum [in/out] Acumulador de claves.
 */
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

/**
 * @brief Cuenta cuántos DNIs empiezan por 8.
 * @return Número de DNIs en el rango [80000000, 89999999].
 */
template<typename TipoClave, typename TipoDato>
unsigned ABB<TipoClave, TipoDato>::cuentaDNI8() const
{
    return cuentaDNI8Aux(miRaiz);
}

/**
 * @brief Auxiliar recursivo para contar DNIs que empiezan por 8.
 * Utiliza poda para no recorrer subárboles innecesarios.
 * @param nodo Nodo actual.
 * @return Cantidad encontrada.
 */
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
