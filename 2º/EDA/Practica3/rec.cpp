#include <iostream>
#include "Pila.h"
#include "error.h"
using namespace std;

int altura(Pila p);
int altura2(Pila& p);
int veces(Pila& p, int x);
int sumaPares(Pila& p);
Pila borrar(Pila p, int x, unsigned n);

/**
 * @brief Programa principal que prueba las funciones recursivas con pilas.
 */
int main()
{
    try
    {
        Pila p(10);
        for (int i = 1; i <= 5; i++)
            p.Apilar(i % 3 + 1);

        cout << "Pila original:\n";
        p.mostrar();

        cout << "Altura (por valor): " << altura(p) << endl;
        cout << "Altura (por referencia): " << altura2(p) << endl;
        cout << "Veces que aparece 2: " << veces(p, 2) << endl;
        cout << "Suma posiciones pares: " << sumaPares(p) << endl;

        Pila borrada = borrar(p, 2, 2);
        cout << "\nPila original: ";
        p.mostrar();
        cout << "Pila borrada: ";
        borrada.mostrar();
    }
    catch (const PilaVacia& e) 
    {
        cout << "Excepcion: " << e.what() << endl;
    }
    catch (const PilaLlena& e) 
    {
        cout << "Excepcion: " << e.what() << endl;
    }

    return 0;
}

/**
 * @brief Calcula la altura de una pila (número de elementos) pasando por valor.
 * @param p Pila pasada por valor.
 * @return Número de elementos de la pila.
 */
int altura(Pila p)
{
    int res;
    if (p.Vacia())
        res = 0;
    else {
        p.Desapilar();
        res = 1 + altura(p);
    }
    return res;
}

/**
 * @brief Calcula la altura de una pila (número de elementos) pasando por referencia.
 * @param p Referencia a la pila.
 * @return Número de elementos de la pila.
 */
int altura2(Pila& p)
{
    int res, x;
    if (p.Vacia())
        res = 0;
    else {
        x = p.Cima();
        p.Desapilar();
        res = 1 + altura2(p);
        p.Apilar(x);
    }
    return res;
}

/**
 * @brief Devuelve cuántas veces aparece un elemento en la pila.
 * @param p Referencia a la pila.
 * @param x Valor a buscar.
 * @return Número de veces que aparece x.
 */
int veces(Pila& p, int x)
{
    int res, cima;
    if (p.Vacia())
        res = 0;
    else 
    {
        cima = p.Cima();
        p.Desapilar();
        res = veces(p, x);
        if (cima == x)
            res++;
        p.Apilar(cima);
    }
    return res;
}

/**
 * @brief Suma los elementos que se encuentran en posiciones pares de la pila.
 * @param p Referencia a la pila.
 * @return Suma de los elementos en posiciones pares.
 */
int sumaPares(Pila& p)
{
    static int pos = 1;
    int res, cima;
    if (p.Vacia()) 
    {
        pos = 1;
        res = 0;
    }
    else {
        cima = p.Cima();
        p.Desapilar();
        res = sumaPares(p);
        if (pos % 2 == 0)
            res += cima;
        p.Apilar(cima);
        pos++;
    }
    return res;
}

/**
 * @brief Devuelve una pila sin las n primeras apariciones de un elemento dado.
 * @param p Pila original.
 * @param x Elemento a eliminar.
 * @param n Número de ocurrencias a borrar.
 * @return Nueva pila sin las n primeras apariciones de x.
 */
Pila borrar(Pila p, int x, unsigned n)
{
    Pila res;
    if (p.Vacia())
        res = p;
    else {
        int cima = p.Cima();
        p.Desapilar();
        res = borrar(p, x, n - (cima == x && n > 0));
        if (cima != x || n == 0)
            res.Apilar(cima);
    }
    return res;
}
