/**
 * @file main.cpp
 * @brief Programa principal que usa Módulos de C++20
 */

#include <iostream>
#include <vector>
#include <cmath>

//Importa los modulos de complejo.cpp
import complejo; 

using namespace std;

/**
 * Genera un vector de números complejos
 */
vector<Complejo> genVector(const Complejo c, int n){
    vector<Complejo> v(n);
    v[0].SetRec(1.0,0.0);
    for(int i = 1; i < n; i++){
        v[i] = v[i-1]* c;
    }
    return v;
}

int main(){

    Complejo c1;
    Complejo c2;

    c1.SetRec(3.0, 4.0);
    c2.SetRec(2.0, 3.0);

    //Prueba suma
    Complejo suma = c1 + c2;
    cout << "Suma: "; 
    c1.Print(); cout<< "+ "; c2.Print(); cout<< "= "; suma.Print();

    //Prueba multiplicación
    Complejo multiplicacion = c1 * c2;
    cout << "Multiplicacion: "; 
    c1.Print(); cout << "* "; c2.Print(); cout << "= "; multiplicacion.Print();

    //Prueba Vector
    Complejo c(2, M_PI/4);
    int n = 10;

    cout << "\nGenerando vector..." << endl;
    vector<Complejo> nuevoVector = genVector(c,n);

    for(size_t i = 0; i < nuevoVector.size(); i++)
    {
        cout << "v[" << i << "] = ";
        nuevoVector[i].Print();
    }

    return 0;
}