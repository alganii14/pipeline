import React, { useState, useEffect } from 'react';
import { getPipelines, createPipeline, updatePipeline, deletePipeline, deleteAllPipelines } from '../api';
import { Plus, Edit2, Trash2, Search, X, Trash } from 'lucide-react';

const Pipelines = () => {
  const [pipelines, setPipelines] = useState([]);
  const [loading, setLoading] = useState(true);
  const [page, setPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [total, setTotal] = useState(0);
  const [search, setSearch] = useState('');
  const [strategy, setStrategy] = useState('');
  const [segment, setSegment] = useState('');
  const [showModal, setShowModal] = useState(false);
  const [modalMode, setModalMode] = useState('create'); // 'create' or 'edit'
  const [currentPipeline, setCurrentPipeline] = useState(null);

  // RFMT search modal states
  const [showRFMTModal, setShowRFMTModal] = useState(false);
  const [rfmtSearch, setRfmtSearch] = useState('');
  const [rfmts, setRfmts] = useState([]);
  const [selectedRFMT, setSelectedRFMT] = useState(null);

  const emptyForm = {
    pn: '',
    nama_rmft: '',
    kode_uker: '',
    kc: '',
    prod: '',
    no_rek: '',
    dup: '',
    nama: '',
    tgl: '',
    strategy: '',
    segment: '',
    pipeline: '',
    proyeksi: 0,
  };

  const [formData, setFormData] = useState(emptyForm);

  useEffect(() => {
    fetchPipelines();
  }, [page, search, strategy, segment]);

  useEffect(() => {
    if (showRFMTModal && rfmtSearch) {
      const timer = setTimeout(() => {
        searchRFMTs();
      }, 300);
      return () => clearTimeout(timer);
    }
  }, [rfmtSearch, showRFMTModal]);

  const fetchPipelines = async () => {
    setLoading(true);
    try {
      const response = await getPipelines({
        page,
        per_page: 10,
        search,
        strategy,
        segment,
      });
      setPipelines(response.data.data);
      setTotalPages(response.data.total_pages);
      setTotal(response.data.total);
      setLoading(false);
    } catch (error) {
      console.error('Error fetching pipelines:', error);
      setLoading(false);
    }
  };

  const searchRFMTs = async () => {
    try {
      const response = await fetch(`http://localhost:8080/api/pipelines/search-rfmts?search=${rfmtSearch}&limit=20`);
      const data = await response.json();
      setRfmts(data || []);
    } catch (error) {
      console.error('Error searching RFMTs:', error);
    }
  };

  const handleSelectRFMT = (rfmt) => {
    setSelectedRFMT(rfmt);
    setFormData({
      ...formData,
      pn: rfmt.pn,
      nama_rmft: rfmt.nama_lengkap
    });
    setShowRFMTModal(false);
    setRfmtSearch('');
  };

  const handleSearch = (e) => {
    setSearch(e.target.value);
    setPage(1);
  };

  const handleFilterChange = (filterType, value) => {
    if (filterType === 'strategy') {
      setStrategy(value);
    } else if (filterType === 'segment') {
      setSegment(value);
    }
    setPage(1);
  };

  const openCreateModal = () => {
    setModalMode('create');
    setFormData(emptyForm);
    setShowModal(true);
  };

  const openEditModal = (pipeline) => {
    setModalMode('edit');
    setCurrentPipeline(pipeline);
    setFormData({
      pn: pipeline.pn,
      nama_rmft: pipeline.nama_rmft,
      kode_uker: pipeline.kode_uker,
      kc: pipeline.kc,
      prod: pipeline.prod,
      no_rek: pipeline.no_rek,
      dup: pipeline.dup,
      nama: pipeline.nama,
      tgl: pipeline.tgl,
      strategy: pipeline.strategy,
      segment: pipeline.segment,
      pipeline: pipeline.pipeline,
      proyeksi: pipeline.proyeksi,
    });
    setShowModal(true);
  };

  const closeModal = () => {
    setShowModal(false);
    setFormData(emptyForm);
    setCurrentPipeline(null);
  };

  const handleInputChange = (e) => {
    const { name, value } = e.target;
    setFormData({ ...formData, [name]: value });
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    try {
      if (modalMode === 'create') {
        await createPipeline(formData);
      } else {
        await updatePipeline(currentPipeline.id, formData);
      }
      closeModal();
      fetchPipelines();
    } catch (error) {
      console.error('Error saving pipeline:', error);
      alert('Failed to save pipeline');
    }
  };

  const handleDelete = async (id) => {
    if (window.confirm('Are you sure you want to delete this pipeline?')) {
      try {
        await deletePipeline(id);
        fetchPipelines();
      } catch (error) {
        console.error('Error deleting pipeline:', error);
        alert('Failed to delete pipeline');
      }
    }
  };

  const handleDeleteAll = async () => {
    if (window.confirm('⚠️ WARNING: This will DELETE ALL pipelines permanently! Are you absolutely sure?')) {
      if (window.confirm('This action CANNOT be undone! Type confirmation is required. Continue?')) {
        try {
          const response = await deleteAllPipelines();
          alert(`✅ ${response.data.message}\nDeleted: ${response.data.deleted_count} pipelines`);
          fetchPipelines();
        } catch (error) {
          console.error('Error deleting all pipelines:', error);
          alert('Failed to delete all pipelines');
        }
      }
    }
  };

  const formatCurrency = (value) => {
    return new Intl.NumberFormat('id-ID', {
      style: 'currency',
      currency: 'IDR',
      minimumFractionDigits: 0,
    }).format(value);
  };

  return (
    <div>
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-3xl font-bold text-gray-800">Pipelines</h1>
        <div className="flex gap-3">
          <button
            onClick={handleDeleteAll}
            className="bg-red-600 hover:bg-red-700 text-white px-4 py-2 rounded-lg flex items-center gap-2 transition"
          >
            <Trash size={20} />
            Delete All
          </button>
          <button
            onClick={openCreateModal}
            className="bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-lg flex items-center gap-2 transition"
          >
            <Plus size={20} />
            Add Pipeline
          </button>
        </div>
      </div>

      {/* Filters */}
      <div className="bg-white rounded-lg shadow-md p-4 mb-6">
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div className="relative">
            <Search className="absolute left-3 top-3 text-gray-400" size={20} />
            <input
              type="text"
              placeholder="Search by PN, Name, or Code..."
              value={search}
              onChange={handleSearch}
              className="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>
          <select
            value={strategy}
            onChange={(e) => handleFilterChange('strategy', e.target.value)}
            className="px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
            <option value="">All Strategies</option>
            <option value="Kolaborasi Perusahaan Anak">Kolaborasi Perusahaan Anak</option>
            <option value="Optimalisasi Business Cluster">Optimalisasi Business Cluster</option>
            <option value="Optimalisasi Digital Channel">Optimalisasi Digital Channel</option>
            <option value="Optimalisasi Nasabah Prio, BOD, BOC">Optimalisasi Nasabah Prio, BOD, BOC</option>
            <option value="Penguatan Produk & Fungsi RM">Penguatan Produk & Fungsi RM</option>
            <option value="Peningkatan Payroll Berkualitas">Peningkatan Payroll Berkualitas</option>
            <option value="Reaktivasi Rek Dormant">Reaktivasi Rek Dormant</option>
            <option value="Rekening Trx Debitur">Rekening Trx Debitur</option>
          </select>
          <select
            value={segment}
            onChange={(e) => handleFilterChange('segment', e.target.value)}
            className="px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
            <option value="">All Segments</option>
            <option value="KONSUMER">KONSUMER</option>
            <option value="Merchant">Merchant</option>
            <option value="Mikro">Mikro</option>
            <option value="Prioritas">Prioritas</option>
            <option value="RITEL BADAN USAHA">RITEL BADAN USAHA</option>
            <option value="RITEL INDIVIDU">RITEL INDIVIDU</option>
            <option value="RITEL NON INDIVIDU">RITEL NON INDIVIDU</option>
            <option value="Ritel Perusahaan">Ritel Perusahaan</option>
            <option value="SME">SME</option>
            <option value="WEALTH">WEALTH</option>
          </select>
        </div>
      </div>

      {/* Table */}
      <div className="bg-white rounded-lg shadow-md overflow-hidden">
        {loading ? (
          <div className="flex items-center justify-center p-12">
            <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
          </div>
        ) : (
          <>
            <div className="overflow-x-auto">
              <div className="overflow-x-auto">
                <table className="min-w-full divide-y divide-gray-200">
                  <thead className="bg-gray-50">
                    <tr>
                      <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">PN</th>
                      <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Nama RMFT</th>
                      <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Kode Uker</th>
                      <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">KC</th>
                      <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Prod</th>
                      <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">No Rek</th>
                      <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Dup</th>
                      <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Nama</th>
                      <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">TGL</th>
                      <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Strategy</th>
                      <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Segment</th>
                      <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Pipeline</th>
                      <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Proyeksi</th>
                      <th className="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider sticky right-0 bg-gray-50">Actions</th>
                    </tr>
                  </thead>
                  <tbody className="bg-white divide-y divide-gray-200">
                    {pipelines.map((pipeline) => (
                      <tr key={pipeline.id} className="hover:bg-gray-50">
                        <td className="px-4 py-3 whitespace-nowrap text-sm font-medium text-gray-900">{pipeline.pn}</td>
                        <td className="px-4 py-3 whitespace-nowrap text-sm text-gray-500">{pipeline.nama_rmft}</td>
                        <td className="px-4 py-3 whitespace-nowrap text-sm text-gray-500">{pipeline.kode_uker}</td>
                        <td className="px-4 py-3 whitespace-nowrap text-sm text-gray-500">{pipeline.kc}</td>
                        <td className="px-4 py-3 whitespace-nowrap text-sm text-gray-500">{pipeline.prod}</td>
                        <td className="px-4 py-3 whitespace-nowrap text-sm text-gray-500">{pipeline.no_rek}</td>
                        <td className="px-4 py-3 whitespace-nowrap text-sm text-gray-500">{pipeline.dup}</td>
                        <td className="px-4 py-3 whitespace-nowrap text-sm text-gray-500">{pipeline.nama}</td>
                        <td className="px-4 py-3 whitespace-nowrap text-sm text-gray-500">{pipeline.tgl}</td>
                        <td className="px-4 py-3 whitespace-nowrap">
                          <span className="px-2 py-1 text-xs rounded-full bg-blue-100 text-blue-800">
                            {pipeline.strategy}
                          </span>
                        </td>
                        <td className="px-4 py-3 whitespace-nowrap">
                          <span className="px-2 py-1 text-xs rounded-full bg-green-100 text-green-800">
                            {pipeline.segment}
                          </span>
                        </td>
                        <td className="px-4 py-3 whitespace-nowrap text-sm text-gray-500">{pipeline.pipeline}</td>
                        <td className="px-4 py-3 whitespace-nowrap text-sm text-gray-500">{formatCurrency(pipeline.proyeksi)}</td>
                        <td className="px-4 py-3 whitespace-nowrap text-right text-sm font-medium sticky right-0 bg-white">
                          <button
                            onClick={() => openEditModal(pipeline)}
                            className="text-blue-600 hover:text-blue-900 mr-3"
                          >
                            <Edit2 size={18} />
                          </button>
                          <button
                            onClick={() => handleDelete(pipeline.id)}
                            className="text-red-600 hover:text-red-900"
                          >
                            <Trash2 size={18} />
                          </button>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>

            {/* Pagination */}
            <div className="bg-gray-50 px-6 py-3 flex items-center justify-between border-t border-gray-200">
              <div className="text-sm text-gray-700">
                Showing <span className="font-medium">{((page - 1) * 10) + 1}</span> to{' '}
                <span className="font-medium">{Math.min(page * 10, total)}</span> of{' '}
                <span className="font-medium">{total}</span> results
              </div>
              <div className="flex gap-2">
                <button
                  onClick={() => setPage(page - 1)}
                  disabled={page === 1}
                  className="px-4 py-2 border border-gray-300 rounded-lg disabled:opacity-50 disabled:cursor-not-allowed hover:bg-gray-100"
                >
                  Previous
                </button>
                <span className="px-4 py-2 border border-gray-300 rounded-lg bg-blue-50 text-blue-600 font-medium">
                  {page}
                </span>
                <button
                  onClick={() => setPage(page + 1)}
                  disabled={page >= totalPages}
                  className="px-4 py-2 border border-gray-300 rounded-lg disabled:opacity-50 disabled:cursor-not-allowed hover:bg-gray-100"
                >
                  Next
                </button>
              </div>
            </div>
          </>
        )}
      </div>

      {/* Modal */}
      {showModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
          <div className="bg-white rounded-lg shadow-xl max-w-2xl w-full max-h-[90vh] overflow-y-auto">
            <div className="flex justify-between items-center p-6 border-b">
              <h2 className="text-2xl font-bold text-gray-800">
                {modalMode === 'create' ? 'Add New Pipeline' : 'Edit Pipeline'}
              </h2>
              <button onClick={closeModal} className="text-gray-500 hover:text-gray-700">
                <X size={24} />
              </button>
            </div>
            <form onSubmit={handleSubmit} className="p-6">
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">PN</label>
                  <input
                    type="text"
                    name="pn"
                    value={formData.pn}
                    readOnly
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg bg-gray-50 focus:outline-none"
                    required
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Nama RMFT</label>
                  <div className="flex gap-2">
                    <input
                      type="text"
                      name="nama_rmft"
                      value={formData.nama_rmft}
                      readOnly
                      className="flex-1 px-3 py-2 border border-gray-300 rounded-lg bg-gray-50 focus:outline-none"
                      required
                    />
                    <button
                      type="button"
                      onClick={() => {
                        setShowRFMTModal(true);
                        searchRFMTs();
                      }}
                      className="px-4 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700 transition whitespace-nowrap"
                    >
                      Select RFMT
                    </button>
                  </div>
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Kode Uker</label>
                  <input
                    type="text"
                    name="kode_uker"
                    value={formData.kode_uker}
                    onChange={handleInputChange}
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                    required
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">KC</label>
                  <input
                    type="text"
                    name="kc"
                    value={formData.kc}
                    onChange={handleInputChange}
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                    required
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Prod</label>
                  <input
                    type="text"
                    name="prod"
                    value={formData.prod}
                    onChange={handleInputChange}
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">No Rek</label>
                  <input
                    type="text"
                    name="no_rek"
                    value={formData.no_rek}
                    onChange={handleInputChange}
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Dup</label>
                  <input
                    type="text"
                    name="dup"
                    value={formData.dup}
                    onChange={handleInputChange}
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Nama</label>
                  <input
                    type="text"
                    name="nama"
                    value={formData.nama}
                    onChange={handleInputChange}
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                    required
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">TGL</label>
                  <input
                    type="date"
                    name="tgl"
                    value={formData.tgl}
                    onChange={handleInputChange}
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Strategy</label>
                  <select
                    name="strategy"
                    value={formData.strategy}
                    onChange={handleInputChange}
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                    required
                  >
                    <option value="">Select Strategy</option>
                    <option value="Kolaborasi Perusahaan Anak">Kolaborasi Perusahaan Anak</option>
                    <option value="Optimalisasi Business Cluster">Optimalisasi Business Cluster</option>
                    <option value="Optimalisasi Digital Channel">Optimalisasi Digital Channel</option>
                    <option value="Optimalisasi Nasabah Prio, BOD, BOC">Optimalisasi Nasabah Prio, BOD, BOC</option>
                    <option value="Penguatan Produk & Fungsi RM">Penguatan Produk & Fungsi RM</option>
                    <option value="Peningkatan Payroll Berkualitas">Peningkatan Payroll Berkualitas</option>
                    <option value="Reaktivasi Rek Dormant">Reaktivasi Rek Dormant</option>
                    <option value="Rekening Trx Debitur">Rekening Trx Debitur</option>
                  </select>
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Segment</label>
                  <select
                    name="segment"
                    value={formData.segment}
                    onChange={handleInputChange}
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                    required
                  >
                    <option value="">Select Segment</option>
                    <option value="KONSUMER">KONSUMER</option>
                    <option value="Merchant">Merchant</option>
                    <option value="Mikro">Mikro</option>
                    <option value="Prioritas">Prioritas</option>
                    <option value="RITEL BADAN USAHA">RITEL BADAN USAHA</option>
                    <option value="RITEL INDIVIDU">RITEL INDIVIDU</option>
                    <option value="RITEL NON INDIVIDU">RITEL NON INDIVIDU</option>
                    <option value="Ritel Perusahaan">Ritel Perusahaan</option>
                    <option value="SME">SME</option>
                    <option value="WEALTH">WEALTH</option>
                  </select>
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Pipeline</label>
                  <input
                    type="text"
                    name="pipeline"
                    value={formData.pipeline}
                    onChange={handleInputChange}
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Proyeksi</label>
                  <input
                    type="number"
                    step="0.01"
                    name="proyeksi"
                    value={formData.proyeksi}
                    onChange={handleInputChange}
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                    required
                  />
                </div>
              </div>
              <div className="flex justify-end gap-3 mt-6">
                <button
                  type="button"
                  onClick={closeModal}
                  className="px-6 py-2 border border-gray-300 rounded-lg hover:bg-gray-100 transition"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition"
                >
                  {modalMode === 'create' ? 'Create' : 'Update'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* RFMT Selection Modal */}
      {showRFMTModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg shadow-xl w-11/12 max-w-3xl max-h-[90vh] overflow-hidden">
            <div className="px-6 py-4 border-b border-gray-200 flex justify-between items-center">
              <h3 className="text-lg font-semibold text-gray-900">Select RFMT</h3>
              <button
                onClick={() => {
                  setShowRFMTModal(false);
                  setRfmtSearch('');
                  setRfmts([]);
                }}
                className="text-gray-400 hover:text-gray-600"
              >
                <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>
            <div className="p-6">
              <div className="mb-4">
                <input
                  type="text"
                  placeholder="Search by PN or Nama Lengkap..."
                  value={rfmtSearch}
                  onChange={(e) => setRfmtSearch(e.target.value)}
                  className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
              </div>
              <div className="overflow-y-auto max-h-96">
                <table className="min-w-full divide-y divide-gray-200">
                  <thead className="bg-gray-50 sticky top-0">
                    <tr>
                      <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">PN</th>
                      <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Nama Lengkap</th>
                      <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">JG</th>
                      <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Action</th>
                    </tr>
                  </thead>
                  <tbody className="bg-white divide-y divide-gray-200">
                    {rfmts.length === 0 ? (
                      <tr>
                        <td colSpan="4" className="px-4 py-8 text-center text-gray-500">
                          {rfmtSearch ? 'No RFMTs found' : 'Type to search RFMTs'}
                        </td>
                      </tr>
                    ) : (
                      rfmts.map((rfmt) => (
                        <tr key={rfmt.id} className="hover:bg-gray-50">
                          <td className="px-4 py-3 whitespace-nowrap text-sm text-gray-900">{rfmt.pn}</td>
                          <td className="px-4 py-3 whitespace-nowrap text-sm text-gray-900">{rfmt.nama_lengkap}</td>
                          <td className="px-4 py-3 whitespace-nowrap text-sm text-gray-500">{rfmt.jg}</td>
                          <td className="px-4 py-3 whitespace-nowrap text-sm">
                            <button
                              onClick={() => handleSelectRFMT(rfmt)}
                              className="px-3 py-1 bg-blue-600 text-white rounded hover:bg-blue-700 transition"
                            >
                              Select
                            </button>
                          </td>
                        </tr>
                      ))
                    )}
                  </tbody>
                </table>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default Pipelines;
