#include "Pila.h"
#include <iostream>
using namespace std;

/**
 * @brief Programa de prueba de la clase Pila.
 *
 * Este programa crea una pila, realiza operaciones de apilado y desapilado,
 * muestra su contenido y verifica el manejo de excepciones.
 */
int main()
{
    try 
    {
        cout << "Creando pila de tamanyo 10...\n";
        Pila p(10);

        cout << "Intentando desapilar en pila vacia...\n";
        try 
        {
            p.Desapilar();
        }
        catch (const exception& e) 
        {
            cout << e.what() << endl;
        }

        cout << "Apilando del 1 al 11...\n";
        for (int i = 1; i <= 11; i++) 
        {
            try 
            {
                p.Apilar(i);
            }
            catch (const exception& e) 
            {
                cout << e.what() << endl;
            }
        }

        cout << "Mostrando pila...\n";
        p.mostrar();

        cout << "Numero de accesos realizados: " << p.getAccesos() << endl;

        cout << "Copiando pila en otra...\n";
        Pila p2 = p;
        cout << "Pila copiada:\n";
        p2.mostrar();
    }
    catch (...) 
    {
        cout << "Se ha producido un error inesperado.\n";
    }

    return 0;
}
