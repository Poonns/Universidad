#include <iostream>
#include <fstream>
#include <cstdlib>
#include <ctime>
#include "Grafo.h"
#include "Prim.h"

using namespace std;

int main() 
{
    srand(time(0)); 
    ResultadoPrim resultado;

    cout << "--- INICIO DE PRUEBAS ---" << endl;

    Grafo g(8);

    try 
    {
        g(0, 1) = 5.0; 
        g(8, 0); 
    } 
    catch (const out_of_range& e) 
    {
        cout << "Excepcion capturada correctamente: " << e.what() << endl;
    }

    g.creaGrafoND(0.5f, 3.0f);
    
    cout << "Matriz de adyacencia:" << endl;
    g.mostrar();

    resultado = prim(g);

    cout << "Arcos del Arbol de Expansion Minima:" << endl;
    for (auto it = resultado.arcos.begin(); it != resultado.arcos.end(); ++it) 
    {
        cout << "(" << it->first << "," << it->second 
             << ") " << g(it->first, it->second) << endl;
    }
    cout << "Total cable: " << resultado.coste_total << " Km" << endl;


    return 0;
}