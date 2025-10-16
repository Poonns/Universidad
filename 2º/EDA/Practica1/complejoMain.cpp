/**
 * @file complejoMain.cpp
 * @brief Programa principal de números complejos
 * @details
 * Comprobar el funcionamiento de los operadores suma y producto de la clase Complejo,
 * además genera un vector de números complejos usando'genVector()' y muestra el contenido del vector
 * 
 */

#include <iostream>
#include <vector>
#include "complejo.h"
using namespace std;


/**
 * @brief Genera un vector de números complejos que cumple la siguiente especificación:
 * Func genVector(c: Complejo; n: Nat) dev v: Vector[0..n-1] de Complejo
 * {Pre: n > 0}
 * {Post: v[0] = (1,0) ∧ ∀α: 0 < α ≤ n-1: v[α] = v[α - 1] * c }
 * 
 * @param c numero complejo
 * @param n tamaño del vector (n>0)
 * @return vector<Complejo> vector generado
 * 
 */
 
 /*
vector<Complejo> genVector(const Complejo c, int n){
    //Creamos vector tamaño n;
    vector<Complejo> v(n);

    //Modificamos v[0] para cumplir con el enunciado
    v[0].SetRec(1.0,0.0);

    //Asginamos los valores al resto de elementos del vector
    for(int i = 1; i < n; i++){
        v[i] = v[i-1]* c;
    }

    return v;
}*/

int main(){
    /**
     * @brief Entrada del programa
     * 
     * Operaciones que realiza:
     *  - Comprueba los operadores de suma y producto de la clase Complejo
     *  - Genera un vector de tamaño n = 10
     *  - Muestra los resultados por pantalla
     */
    
    Complejo c1;
    Complejo c2;

    //Asignamos valores a los objetos creados
    c1.SetRec(3.0, 4.0);
    c2.SetRec(2.0, 3.0);

    //Prueba del operator+
    Complejo suma = c1 + c2;
    cout << "Suma: "; 
    c1.Print();
    cout<< "+"; 
    c2.Print();
    cout<< "="; 
    suma.Print();

    //Prueba del operator*
    Complejo multiplicacion = c1 * c2;
    cout << "Multiplicación: "; 
    c1.Print();
    cout << "*"; 
    c2.Print(); 
    cout << "= "; 
    multiplicacion.Print();

    //Asignamos valores a los parámetros que pasaremos a la función definida previamente
    Complejo c(2,M_PI/4);
/*    int n = 10;

    //Creamos el vector y lo mostramos por pantalla
    vector<Complejo> nuevoVector = genVector(c,n);

    for(int i = 0; i < nuevoVector.size(); i++)
    {
        cout << "v[" << i << "] = ";
        nuevoVector[i].print();
    }
*/
    return 0;
}