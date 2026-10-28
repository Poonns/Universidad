#include <iostream>
#include <fstream>
#include <cstdlib>
#include <ctime>
#include "Grafo.h"
#include "Prim.h"

using namespace std;

int main (void)
{
    srand(time(0));
    ofstream f;
    string nombre = "costes.dat";
    double media_accesos;
    long total_accesos;
    int repeticiones;
    f.open(nombre);

    if (!f)
        cout << "Error abriendo el fichero" << endl; 

    cout << "Calculando costes medios..." << endl; 
    cout << "Accesos" << endl; 

    for  (unsigned n = 5; n <= 100; n+= 5)
    {
        total_accesos = 0;
        repeticiones = 10;

        for (int r = 0; r < repeticiones; ++r)
        {
            Grafo g(n);
            g.creaGrafoND(0.5f, 3.0f);
            g.resetAccesos();

            prim(g);

            total_accesos += g.getAccesos();
        }
        media_accesos = (double)total_accesos / repeticiones;

        cout << n << "\t" << media_accesos << endl;
        
        f << n << "," << media_accesos << endl; 
    }

    f.close();

    /* CONCLUSIONES:
     * El algoritmo de Prim con Matriz de Adyacencia tiene una complejidad teorica de O(n^2).
     * La grafica resultante muestra un ajuste cuadratico perfecto sobre los datos empiricos, 
     * confirmando esta complejidad teorica.
    */

    return 0;
}