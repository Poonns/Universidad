//MÉTODO QUE BUSCA SI EXISTE UN ELEMENTO
bool Tabla::buscar(TipoClave cl, TipoDato &dato) const{
    unsigned pos = hash(cl) % t.size();

    for(int i = 0; i < t[pos].size(); i++){
        if(t[pos][i].clave == cl){
            dato = t[pos][i].dato;
            return true;
        }
    }
    return false;
}