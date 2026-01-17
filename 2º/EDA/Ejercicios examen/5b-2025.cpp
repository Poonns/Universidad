//MËTODO QUE TE CUENTA CUANTOS NODOS SE ENCUENTRAN EN UN RANGO DE VALOR

unsigned ABB::contarEnRangoAux(NodoBin* nodo, int min, int max) const{
    if(nodo == nullptr){
        return 0;
    }

    if(nodo->dato < min){
        contarEnRangoAux(nodo->der, min, max);
    }

    if(nodo->dato > max){
        contarEnRangoAux(nodo->izq, min, max);
    }

    return 1 + contarEnrangoAux(nodo->der, min, max) + contarEnRangoAux(nodo->izq, min, max);
}