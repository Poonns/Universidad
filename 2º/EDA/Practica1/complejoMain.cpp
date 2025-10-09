#include <iostream>
#include <vector>
#include "complejo.cpp"
using namespace std;


/**
 * Declaramos la función que nos añadirá los complejos al vector y cumplira lo siguiente:
 * 
 * Func genVector(c: Complejo; n: Nat) dev v: Vector[0..n-1] de Complejo
 * {Pre: n > 0}
 * {Post: v[0] = (1,0) ∧ ∀α: 0 < α ≤ n-1: v[α] = v[α - 1] * c }
 */
vector<Complejo> genVector(const Complejo c, int n){
    //Creamos vector tamaño n;
    vector<Complejo> v(n);

    //Modificamos v[0] para cumplir con el enunciado
    v[0].setRec(1,0);

    //Asginamos los valores al resto de elementos del vector
    for(int i = 1; i < n; i++){
        v[i] = v[i-1]* c;
    }

    return v;
}

int main(){
    /**
     * Creamos dos objetos para comprobar si las 
     * operaciones se realizan de manera correcta
     */
    
    Complejo c1;
    Complejo c2;

    //Asignamos valores a los objetos creados
    c1.setRec(3,4);
    c2.setRec(2,3);

    //Prueba del operator+
    Complejo suma = c1 + c2;
    cout << "Suma: "; 
    c1.print();
    cout<< "+"; 
    c2.print();
    cout<< "="; 
    suma.print();

    //Prueba del operator*
    Complejo multiplicacion = c1 * c2;
    cout << "Multiplicación: "; 
    c1.print();
    cout << "*"; 
    c2.print(); 
    cout << "= "; 
    multiplicacion.print();

    //Asignamos valores a los parámetros que pasaremos a la función definida previamente
    Complejo c(2,M_PI/4);
    int n = 10;

    //Creamos el vector y lo mostramos por pantalla
    vector<Complejo> nuevoVector = genVector(c,n);

    for(int i = 0; i < nuevoVector.size(); i++){
        cout << "v[" << i << "] = ";
        nuevoVector[i].print();
    }

    return 0;
}